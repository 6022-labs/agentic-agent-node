package middlewares_test

import (
	"encoding/base64"
	"encoding/json"
	"errors"
	"io"
	"strings"
	"testing"

	"github.com/6022-labs/agentic-agent-node/src/gateway/payment/models"
	"github.com/6022-labs/agentic-agent-node/src/gateway/payment/x402"
	x402_interfaces "github.com/6022-labs/agentic-agent-node/src/gateway/payment/x402/services/interfaces"
	core_middlewares "github.com/6022-labs/agentic-agent-node/src/gateway_mvc/core/middlewares"
	"github.com/6022-labs/agentic-agent-node/src/gateway_mvc/reasoning/middlewares"
	middleware_interfaces "github.com/6022-labs/agentic-agent-node/src/gateway_mvc/reasoning/middlewares/interfaces"
	core_interfaces_mocks "github.com/6022-labs/agentic-agent-node/tests/gateway_mocks/core_mocks/services_mocks/interfaces_mocks"
	interfaces_mocks "github.com/6022-labs/agentic-agent-node/tests/gateway_mocks/payment_mocks/services_mocks/interfaces_mocks"
	x402_mocks "github.com/6022-labs/agentic-agent-node/tests/gateway_mocks/payment_mocks/x402_mocks/services_mocks/interfaces_mocks"
	metrics_mocks "github.com/6022-labs/agentic-agent-node/tests/gateway_mvc_mocks/metrics_mocks/interfaces_mocks"
	"github.com/ethereum/go-ethereum/common"
	"github.com/gofiber/fiber/v2"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
	"go.uber.org/zap"
)

const (
	x402TestPayerHex       = "0x70997970C51812dc3A010C7d01b50e0d17dc79C8"
	x402TestAgentWalletHex = "0x3C44CdDdB6a900fa2b585dd299e03d12FA4293BC"
	x402TestPayToHex       = "0xf39Fd6e51aad88F6F4ce6aB8827279cffFb92266"
)

type WhenHandlingX402GateTestingSuite struct {
	sut *middlewares.ResponseAccessGateMiddleware
	app *fiber.App

	handlerCalled bool
	seenPayer     string
	seenAddress   string

	mockPaymentRuleClient        *interfaces_mocks.MockPaymentRuleClientInterface
	mockPaymentFacilitatorClient *interfaces_mocks.MockPaymentFacilitatorClientInterface
	mockJwtValidator             *core_interfaces_mocks.MockJwtValidatorInterface
	mockX402PaymentVerifier      *x402_mocks.MockX402PaymentVerifierInterface
	mockX402FacilitatorClient    *x402_mocks.MockX402FacilitatorClientInterface
}

func WhenHandlingX402GateBeforeEach(t *testing.T) *WhenHandlingX402GateTestingSuite {
	mockController := gomock.NewController(t)

	mockPaymentRuleClient := interfaces_mocks.NewMockPaymentRuleClientInterface(mockController)
	mockPaymentFacilitatorClient := interfaces_mocks.NewMockPaymentFacilitatorClientInterface(mockController)
	mockJwtValidator := core_interfaces_mocks.NewMockJwtValidatorInterface(mockController)
	mockX402PaymentVerifier := x402_mocks.NewMockX402PaymentVerifierInterface(mockController)
	mockX402FacilitatorClient := x402_mocks.NewMockX402FacilitatorClientInterface(mockController)

	mockPaymentMetrics := metrics_mocks.NewMockPaymentMetricsInterface(mockController)
	mockPaymentMetrics.EXPECT().RecordSettled(gomock.Any(), gomock.Any()).AnyTimes()
	mockPaymentMetrics.EXPECT().RecordRejected(gomock.Any(), gomock.Any()).AnyTimes()

	x402Method := middlewares.NewX402AccessMethod(
		zap.NewNop(),
		mockPaymentFacilitatorClient,
		mockX402PaymentVerifier,
		mockX402FacilitatorClient,
		mockPaymentMetrics,
	)
	jwtMethod := middlewares.NewJwtFreeAccessMethod(middlewares.NewJwtFreeAccessMethodParams{
		JwtValidator:     mockJwtValidator,
		FreeRuleMatchers: []middleware_interfaces.FreeRuleMatcherInterface{x402Method},
	})
	sut := middlewares.NewResponseAccessGateMiddleware(middlewares.NewResponseAccessGateMiddlewareParams{
		PaymentRuleClient: mockPaymentRuleClient,
		AccessMethods:     []middleware_interfaces.AccessMethodInterface{x402Method, jwtMethod},
	})

	suite := &WhenHandlingX402GateTestingSuite{
		sut:                          sut,
		mockPaymentRuleClient:        mockPaymentRuleClient,
		mockPaymentFacilitatorClient: mockPaymentFacilitatorClient,
		mockJwtValidator:             mockJwtValidator,
		mockX402PaymentVerifier:      mockX402PaymentVerifier,
		mockX402FacilitatorClient:    mockX402FacilitatorClient,
	}

	app := fiber.New()
	app.Post("/responses", sut.Handle, func(ctx *fiber.Ctx) error {
		suite.handlerCalled = true
		suite.seenPayer, _ = ctx.Locals(middlewares.X402PayerLocalKey).(string)
		suite.seenAddress, _ = ctx.Locals(core_middlewares.AuthAddressLocalKey).(string)
		return ctx.JSON(fiber.Map{"answer": "42"})
	})
	suite.app = app

	return suite
}

func TestWhenHandlingX402Gate(t *testing.T) {
	t.Parallel()

	t.Run("Given fetching the rules fails", func(t *testing.T) {
		t.Parallel()

		initSuite := func(suite *WhenHandlingX402GateTestingSuite) {
			suite.mockPaymentRuleClient.EXPECT().
				List(gomock.Any()).
				Return(nil, errors.New("runtime down")).
				Times(1)
		}

		t.Run("Should return 502", func(t *testing.T) {
			t.Parallel()

			suite := WhenHandlingX402GateBeforeEach(t)
			initSuite(suite)

			resp, err := suite.app.Test(x402GateRequest(""))

			assert.NoError(t, err)
			assert.Equal(t, fiber.StatusBadGateway, resp.StatusCode)
		})
	})

	t.Run("Given fetching the facilitators fails", func(t *testing.T) {
		t.Parallel()

		initSuite := func(suite *WhenHandlingX402GateTestingSuite) {
			suite.mockPaymentRuleClient.EXPECT().
				List(gomock.Any()).
				Return(x402GateTestRules(x402GatePublicRule()), nil).
				Times(1)
			suite.mockPaymentFacilitatorClient.EXPECT().
				Get(gomock.Any()).
				Return(nil, errors.New("runtime down")).
				Times(1)
		}

		t.Run("Should return 502", func(t *testing.T) {
			t.Parallel()

			suite := WhenHandlingX402GateBeforeEach(t)
			initSuite(suite)

			resp, err := suite.app.Test(x402GateRequest(""))

			assert.NoError(t, err)
			assert.Equal(t, fiber.StatusBadGateway, resp.StatusCode)
		})
	})

	t.Run("Given no payment header is sent", func(t *testing.T) {
		t.Parallel()

		initSuite := func(suite *WhenHandlingX402GateTestingSuite) {
			suite.mockPaymentRuleClient.EXPECT().
				List(gomock.Any()).
				Return(x402GateTestRules(
					x402GatePublicRule(),
					x402GateFreeRule(strings.ToLower(x402TestAgentWalletHex)),
				), nil).
				Times(1)
			suite.mockPaymentFacilitatorClient.EXPECT().Get(gomock.Any()).Return(x402GateFacilitators(), nil).Times(1)
		}

		t.Run("Should return 402 advertising only the public rules", func(t *testing.T) {
			t.Parallel()

			suite := WhenHandlingX402GateBeforeEach(t)
			initSuite(suite)

			resp, err := suite.app.Test(x402GateRequest(""))

			assert.NoError(t, err)
			assert.Equal(t, fiber.StatusPaymentRequired, resp.StatusCode)
			assert.False(t, suite.handlerCalled)

			required := decodePaymentRequiredHeader(t, resp)
			assert.Equal(t, x402.Version, required.X402Version)
			// The scoped zero-price rule is a private offer — not listed.
			assert.Len(t, required.Accepts, 1)
			assert.Equal(t, "10000", required.Accepts[0].Amount)
			// The settlement timeout is advertised from the facilitator config.
			assert.Equal(t, 120, required.Accepts[0].MaxTimeoutSeconds)
			// The signing method is advertised so the client knows how to sign;
			// eip3009 also carries the token's EIP-712 domain.
			assert.Equal(t, x402.AssetTransferMethodEip3009, required.Accepts[0].Extra["assetTransferMethod"])
			assert.Equal(t, "USDC", required.Accepts[0].Extra["name"])
			assert.Equal(t, "2", required.Accepts[0].Extra["version"])

			bodyBytes, _ := io.ReadAll(resp.Body)
			var body x402.PaymentRequired
			assert.NoError(t, json.Unmarshal(bodyBytes, &body))
			assert.Len(t, body.Accepts, 1)
		})
	})

	t.Run("Given a public permit2 rule and no payment header", func(t *testing.T) {
		t.Parallel()

		initSuite := func(suite *WhenHandlingX402GateTestingSuite) {
			suite.mockPaymentRuleClient.EXPECT().
				List(gomock.Any()).
				Return(x402GateTestRules(x402GatePermit2Rule()), nil).
				Times(1)
			suite.mockPaymentFacilitatorClient.EXPECT().Get(gomock.Any()).Return(x402GateFacilitators(), nil).Times(1)
		}

		t.Run("Should advertise permit2 with no EIP-712 domain", func(t *testing.T) {
			t.Parallel()

			suite := WhenHandlingX402GateBeforeEach(t)
			initSuite(suite)

			resp, err := suite.app.Test(x402GateRequest(""))

			assert.NoError(t, err)
			assert.Equal(t, fiber.StatusPaymentRequired, resp.StatusCode)

			required := decodePaymentRequiredHeader(t, resp)
			assert.Len(t, required.Accepts, 1)
			extra := required.Accepts[0].Extra
			assert.Equal(t, x402.AssetTransferMethodPermit2, extra["assetTransferMethod"])
			// Permit2 signs over its own fixed domain — no token name/version.
			assert.NotContains(t, extra, "name")
			assert.NotContains(t, extra, "version")
		})
	})

	t.Run("Given a fresh node with only the seeded scoped free rule", func(t *testing.T) {
		t.Parallel()

		initSuite := func(suite *WhenHandlingX402GateTestingSuite) {
			suite.mockPaymentRuleClient.EXPECT().
				List(gomock.Any()).
				Return(x402GateTestRules(
					x402GateFreeRule(strings.ToLower(x402TestAgentWalletHex)),
				), nil).
				Times(1)
			suite.mockPaymentFacilitatorClient.EXPECT().Get(gomock.Any()).Return(x402GateFacilitators(), nil).Times(1)
		}

		t.Run("Should return 402 with an empty offer (closed to outsiders)", func(t *testing.T) {
			t.Parallel()

			suite := WhenHandlingX402GateBeforeEach(t)
			initSuite(suite)

			resp, err := suite.app.Test(x402GateRequest(""))

			assert.NoError(t, err)
			assert.Equal(t, fiber.StatusPaymentRequired, resp.StatusCode)
			assert.Empty(t, decodePaymentRequiredHeader(t, resp).Accepts)
		})
	})

	t.Run("Given the payment header is malformed", func(t *testing.T) {
		t.Parallel()

		initSuite := func(suite *WhenHandlingX402GateTestingSuite) {
			suite.mockPaymentRuleClient.EXPECT().
				List(gomock.Any()).
				Return(x402GateTestRules(x402GatePublicRule()), nil).
				Times(1)
			suite.mockPaymentFacilitatorClient.EXPECT().Get(gomock.Any()).Return(x402GateFacilitators(), nil).Times(1)
		}

		t.Run("Should return 402 with the malformed error code", func(t *testing.T) {
			t.Parallel()

			suite := WhenHandlingX402GateBeforeEach(t)
			initSuite(suite)

			resp, err := suite.app.Test(x402GateRequest("not-base64!!"))

			assert.NoError(t, err)
			assert.Equal(t, fiber.StatusPaymentRequired, resp.StatusCode)
			assert.Equal(t, "malformed_payment_signature", decodePaymentRequiredHeader(t, resp).Error)
		})
	})

	t.Run("Given the payment matches no configured rule", func(t *testing.T) {
		t.Parallel()

		initSuite := func(suite *WhenHandlingX402GateTestingSuite) {
			suite.mockPaymentRuleClient.EXPECT().
				List(gomock.Any()).
				Return(x402GateTestRules(x402GatePublicRule()), nil).
				Times(1)
			suite.mockPaymentFacilitatorClient.EXPECT().Get(gomock.Any()).Return(x402GateFacilitators(), nil).Times(1)
		}

		t.Run("Should return 402 with no matching requirements", func(t *testing.T) {
			t.Parallel()

			suite := WhenHandlingX402GateBeforeEach(t)
			initSuite(suite)

			header := x402GatePaymentHeader(t, "eip155:80002", x402TestPayerHex, "10000")
			resp, err := suite.app.Test(x402GateRequest(header))

			assert.NoError(t, err)
			assert.Equal(t, fiber.StatusPaymentRequired, resp.StatusCode)
			assert.Equal(t, "no_matching_payment_requirements", decodePaymentRequiredHeader(t, resp).Error)
		})
	})

	t.Run("Given the payment signature is invalid", func(t *testing.T) {
		t.Parallel()

		initSuite := func(suite *WhenHandlingX402GateTestingSuite) {
			suite.mockPaymentRuleClient.EXPECT().
				List(gomock.Any()).
				Return(x402GateTestRules(x402GatePublicRule()), nil).
				Times(1)
			suite.mockPaymentFacilitatorClient.EXPECT().Get(gomock.Any()).Return(x402GateFacilitators(), nil).Times(1)

			suite.mockX402PaymentVerifier.EXPECT().
				RecoverPayer(gomock.Any(), gomock.Any()).
				Return(common.Address{}, errors.New("recovered signer does not match")).
				Times(1)
		}

		t.Run("Should return 402 with the invalid signature error code", func(t *testing.T) {
			t.Parallel()

			suite := WhenHandlingX402GateBeforeEach(t)
			initSuite(suite)

			header := x402GatePaymentHeader(t, "eip155:80002", x402TestPayToHex, "10000")
			resp, err := suite.app.Test(x402GateRequest(header))

			assert.NoError(t, err)
			assert.Equal(t, fiber.StatusPaymentRequired, resp.StatusCode)
			assert.Equal(t, "invalid_payment_signature", decodePaymentRequiredHeader(t, resp).Error)
		})
	})

	t.Run("Given the authorization window has expired", func(t *testing.T) {
		t.Parallel()

		initSuite := func(suite *WhenHandlingX402GateTestingSuite) {
			suite.mockPaymentRuleClient.EXPECT().
				List(gomock.Any()).
				Return(x402GateTestRules(x402GatePublicRule()), nil).
				Times(1)
			suite.mockPaymentFacilitatorClient.EXPECT().Get(gomock.Any()).Return(x402GateFacilitators(), nil).Times(1)

			suite.mockX402PaymentVerifier.EXPECT().
				RecoverPayer(gomock.Any(), gomock.Any()).
				Return(common.Address{}, x402_interfaces.ErrX402AuthorizationExpired).
				Times(1)
		}

		t.Run("Should return 402 with the expired error code", func(t *testing.T) {
			t.Parallel()

			suite := WhenHandlingX402GateBeforeEach(t)
			initSuite(suite)

			header := x402GatePaymentHeader(t, "eip155:80002", x402TestPayToHex, "10000")
			resp, err := suite.app.Test(x402GateRequest(header))

			assert.NoError(t, err)
			assert.Equal(t, fiber.StatusPaymentRequired, resp.StatusCode)
			assert.Equal(t, "authorization_expired", decodePaymentRequiredHeader(t, resp).Error)
		})
	})

	t.Run("Given the payer matches a zero-price rule", func(t *testing.T) {
		t.Parallel()

		initSuite := func(suite *WhenHandlingX402GateTestingSuite) {
			suite.mockPaymentRuleClient.EXPECT().
				List(gomock.Any()).
				Return(x402GateTestRules(
					x402GateFreeRule(strings.ToLower(x402TestPayerHex)),
				), nil).
				Times(1)
			suite.mockPaymentFacilitatorClient.EXPECT().Get(gomock.Any()).Return(x402GateFacilitators(), nil).Times(1)

			suite.mockX402PaymentVerifier.EXPECT().
				RecoverPayer(gomock.Any(), gomock.Any()).
				DoAndReturn(func(_ *x402.PaymentPayload, config *x402.X402Config) (common.Address, error) {
					assert.Equal(t, x402TestAgentWalletHex, config.PayTo)
					return common.HexToAddress(x402TestPayerHex), nil
				}).
				Times(1)
		}

		t.Run("Should serve for free without contacting the facilitator", func(t *testing.T) {
			t.Parallel()

			suite := WhenHandlingX402GateBeforeEach(t)
			initSuite(suite)

			header := x402GatePaymentHeader(t, "eip155:80002", x402TestAgentWalletHex, "0")
			resp, err := suite.app.Test(x402GateRequest(header))

			assert.NoError(t, err)
			assert.Equal(t, fiber.StatusOK, resp.StatusCode)
			assert.True(t, suite.handlerCalled)
			assert.Equal(t, common.HexToAddress(x402TestPayerHex).Hex(), suite.seenPayer)
			assert.Equal(t, common.HexToAddress(x402TestPayerHex).Hex(), suite.seenAddress)
			assert.Empty(t, resp.Header.Get(middlewares.PaymentResponseHeader))
		})
	})

	t.Run("Given an everyone-free rule", func(t *testing.T) {
		t.Parallel()

		initSuite := func(suite *WhenHandlingX402GateTestingSuite) {
			suite.mockPaymentRuleClient.EXPECT().
				List(gomock.Any()).
				Return(x402GateTestRules(x402GateFreeRule(".*")), nil).
				Times(1)
			suite.mockPaymentFacilitatorClient.EXPECT().Get(gomock.Any()).Return(x402GateFacilitators(), nil).Times(1)

			suite.mockX402PaymentVerifier.EXPECT().
				RecoverPayer(gomock.Any(), gomock.Any()).
				Return(common.HexToAddress(x402TestPayerHex), nil).
				Times(1)
		}

		t.Run("Should serve any payer for free", func(t *testing.T) {
			t.Parallel()

			suite := WhenHandlingX402GateBeforeEach(t)
			initSuite(suite)

			header := x402GatePaymentHeader(t, "eip155:80002", x402TestAgentWalletHex, "0")
			resp, err := suite.app.Test(x402GateRequest(header))

			assert.NoError(t, err)
			assert.Equal(t, fiber.StatusOK, resp.StatusCode)
			assert.True(t, suite.handlerCalled)
		})
	})

	t.Run("Given a public paid rule has no facilitator", func(t *testing.T) {
		t.Parallel()

		initSuite := func(suite *WhenHandlingX402GateTestingSuite) {
			suite.mockPaymentRuleClient.EXPECT().
				List(gomock.Any()).
				Return(x402GateTestRules(x402GatePublicRule()), nil).
				Times(1)
			// No facilitator configured for the rule's network.
			suite.mockPaymentFacilitatorClient.EXPECT().
				Get(gomock.Any()).
				Return(map[string]models.ExternalPaymentFacilitator{}, nil).
				Times(1)
		}

		t.Run("Should return 402 without advertising the unsettleable rule", func(t *testing.T) {
			t.Parallel()

			suite := WhenHandlingX402GateBeforeEach(t)
			initSuite(suite)

			resp, err := suite.app.Test(x402GateRequest(""))

			assert.NoError(t, err)
			assert.Equal(t, fiber.StatusPaymentRequired, resp.StatusCode)
			assert.Empty(t, decodePaymentRequiredHeader(t, resp).Accepts)
		})
	})

	t.Run("Given a free rule has no facilitator", func(t *testing.T) {
		t.Parallel()

		initSuite := func(suite *WhenHandlingX402GateTestingSuite) {
			suite.mockPaymentRuleClient.EXPECT().
				List(gomock.Any()).
				Return(x402GateTestRules(x402GateFreeRule(strings.ToLower(x402TestPayerHex))), nil).
				Times(1)
			suite.mockPaymentFacilitatorClient.EXPECT().
				Get(gomock.Any()).
				Return(map[string]models.ExternalPaymentFacilitator{}, nil).
				Times(1)

			suite.mockX402PaymentVerifier.EXPECT().
				RecoverPayer(gomock.Any(), gomock.Any()).
				Return(common.HexToAddress(x402TestPayerHex), nil).
				Times(1)
		}

		t.Run("Should still serve it for free", func(t *testing.T) {
			t.Parallel()

			suite := WhenHandlingX402GateBeforeEach(t)
			initSuite(suite)

			header := x402GatePaymentHeader(t, "eip155:80002", x402TestAgentWalletHex, "0")
			resp, err := suite.app.Test(x402GateRequest(header))

			assert.NoError(t, err)
			assert.Equal(t, fiber.StatusOK, resp.StatusCode)
			assert.True(t, suite.handlerCalled)
		})
	})

	t.Run("Given the payer matches only a facilitator-less paid rule", func(t *testing.T) {
		t.Parallel()

		initSuite := func(suite *WhenHandlingX402GateTestingSuite) {
			suite.mockPaymentRuleClient.EXPECT().
				List(gomock.Any()).
				Return(x402GateTestRules(x402GatePublicRule()), nil).
				Times(1)
			suite.mockPaymentFacilitatorClient.EXPECT().
				Get(gomock.Any()).
				Return(map[string]models.ExternalPaymentFacilitator{}, nil).
				Times(1)

			suite.mockX402PaymentVerifier.EXPECT().
				RecoverPayer(gomock.Any(), gomock.Any()).
				Return(common.HexToAddress(x402TestPayerHex), nil).
				Times(1)
		}

		t.Run("Should refuse the payment without contacting the facilitator", func(t *testing.T) {
			t.Parallel()

			suite := WhenHandlingX402GateBeforeEach(t)
			initSuite(suite)

			header := x402GatePaymentHeader(t, "eip155:80002", x402TestPayToHex, "10000")
			resp, err := suite.app.Test(x402GateRequest(header))

			assert.NoError(t, err)
			assert.Equal(t, fiber.StatusPaymentRequired, resp.StatusCode)
			assert.Equal(t, "no_matching_payment_requirements", decodePaymentRequiredHeader(t, resp).Error)
			assert.False(t, suite.handlerCalled)
		})
	})

	t.Run("Given a stranger signs against a scoped rule", func(t *testing.T) {
		t.Parallel()

		initSuite := func(suite *WhenHandlingX402GateTestingSuite) {
			suite.mockPaymentRuleClient.EXPECT().
				List(gomock.Any()).
				Return(x402GateTestRules(
					x402GateFreeRule("0x1111111111111111111111111111111111111111"),
				), nil).
				Times(1)
			suite.mockPaymentFacilitatorClient.EXPECT().Get(gomock.Any()).Return(x402GateFacilitators(), nil).Times(1)

			suite.mockX402PaymentVerifier.EXPECT().
				RecoverPayer(gomock.Any(), gomock.Any()).
				Return(common.HexToAddress(x402TestPayerHex), nil).
				Times(1)
		}

		t.Run("Should return 402 — the rule is not theirs to use", func(t *testing.T) {
			t.Parallel()

			suite := WhenHandlingX402GateBeforeEach(t)
			initSuite(suite)

			header := x402GatePaymentHeader(t, "eip155:80002", x402TestAgentWalletHex, "0")
			resp, err := suite.app.Test(x402GateRequest(header))

			assert.NoError(t, err)
			assert.Equal(t, fiber.StatusPaymentRequired, resp.StatusCode)
			assert.Equal(t, "no_matching_payment_requirements", decodePaymentRequiredHeader(t, resp).Error)
			assert.False(t, suite.handlerCalled)
		})
	})

	t.Run("Given the payer is entitled to a discounted rule", func(t *testing.T) {
		t.Parallel()

		initSuite := func(suite *WhenHandlingX402GateTestingSuite) {
			discounted := x402GateRule("5000", x402TestPayToHex, []string{strings.ToLower(x402TestPayerHex)})

			suite.mockPaymentRuleClient.EXPECT().
				List(gomock.Any()).
				Return(x402GateTestRules(x402GatePublicRule(), discounted), nil).
				Times(1)
			suite.mockPaymentFacilitatorClient.EXPECT().Get(gomock.Any()).Return(x402GateFacilitators(), nil).Times(1)

			suite.mockX402PaymentVerifier.EXPECT().
				RecoverPayer(gomock.Any(), gomock.Any()).
				Return(common.HexToAddress(x402TestPayerHex), nil).
				Times(2)

			suite.mockX402FacilitatorClient.EXPECT().
				Verify(gomock.Any(), "https://x402-amoy.polygon.technology", gomock.Any(), gomock.Any()).
				DoAndReturn(func(_ any, _ string, _ *x402.PaymentPayload, requirements *x402.PaymentRequirements) (*x402.VerifyResponse, error) {
					// The cheapest entitled rule prices the call; the timeout
					// comes from the facilitator config.
					assert.Equal(t, "5000", requirements.Amount)
					assert.Equal(t, 120, requirements.MaxTimeoutSeconds)
					return &x402.VerifyResponse{IsValid: true}, nil
				}).
				Times(1)

			suite.mockX402FacilitatorClient.EXPECT().
				Settle(gomock.Any(), "https://x402-amoy.polygon.technology", gomock.Any(), gomock.Any()).
				Return(&x402.SettleResponse{Success: true, Transaction: "0xtxhash"}, nil).
				Times(1)
		}

		t.Run("Should charge the discounted price", func(t *testing.T) {
			t.Parallel()

			suite := WhenHandlingX402GateBeforeEach(t)
			initSuite(suite)

			header := x402GatePaymentHeader(t, "eip155:80002", x402TestPayToHex, "5000")
			resp, err := suite.app.Test(x402GateRequest(header))

			assert.NoError(t, err)
			assert.Equal(t, fiber.StatusOK, resp.StatusCode)
			assert.True(t, suite.handlerCalled)
		})
	})

	t.Run("Given a paying caller signs less than their price", func(t *testing.T) {
		t.Parallel()

		initSuite := func(suite *WhenHandlingX402GateTestingSuite) {
			suite.mockPaymentRuleClient.EXPECT().
				List(gomock.Any()).
				Return(x402GateTestRules(x402GatePublicRule()), nil).
				Times(1)
			suite.mockPaymentFacilitatorClient.EXPECT().Get(gomock.Any()).Return(x402GateFacilitators(), nil).Times(1)

			suite.mockX402PaymentVerifier.EXPECT().
				RecoverPayer(gomock.Any(), gomock.Any()).
				Return(common.HexToAddress(x402TestPayerHex), nil).
				Times(1)
		}

		t.Run("Should return 402 with the insufficient amount error code", func(t *testing.T) {
			t.Parallel()

			suite := WhenHandlingX402GateBeforeEach(t)
			initSuite(suite)

			header := x402GatePaymentHeader(t, "eip155:80002", x402TestPayToHex, "1")
			resp, err := suite.app.Test(x402GateRequest(header))

			assert.NoError(t, err)
			assert.Equal(t, fiber.StatusPaymentRequired, resp.StatusCode)
			assert.Equal(t, "insufficient_amount", decodePaymentRequiredHeader(t, resp).Error)
		})
	})

	t.Run("Given the facilitator rejects the payment", func(t *testing.T) {
		t.Parallel()

		initSuite := func(suite *WhenHandlingX402GateTestingSuite) {
			suite.mockPaymentRuleClient.EXPECT().
				List(gomock.Any()).
				Return(x402GateTestRules(x402GatePublicRule()), nil).
				Times(1)
			suite.mockPaymentFacilitatorClient.EXPECT().Get(gomock.Any()).Return(x402GateFacilitators(), nil).Times(1)

			suite.mockX402PaymentVerifier.EXPECT().
				RecoverPayer(gomock.Any(), gomock.Any()).
				Return(common.HexToAddress(x402TestPayerHex), nil).
				Times(1)

			suite.mockX402FacilitatorClient.EXPECT().
				Verify(gomock.Any(), "https://x402-amoy.polygon.technology", gomock.Any(), gomock.Any()).
				Return(&x402.VerifyResponse{IsValid: false, InvalidReason: "insufficient_funds"}, nil).
				Times(1)
		}

		t.Run("Should return 402 with the facilitator's reason", func(t *testing.T) {
			t.Parallel()

			suite := WhenHandlingX402GateBeforeEach(t)
			initSuite(suite)

			header := x402GatePaymentHeader(t, "eip155:80002", x402TestPayToHex, "10000")
			resp, err := suite.app.Test(x402GateRequest(header))

			assert.NoError(t, err)
			assert.Equal(t, fiber.StatusPaymentRequired, resp.StatusCode)
			assert.Equal(t, "insufficient_funds", decodePaymentRequiredHeader(t, resp).Error)
			assert.False(t, suite.handlerCalled)
		})
	})

	t.Run("Given the facilitator is unreachable", func(t *testing.T) {
		t.Parallel()

		initSuite := func(suite *WhenHandlingX402GateTestingSuite) {
			suite.mockPaymentRuleClient.EXPECT().
				List(gomock.Any()).
				Return(x402GateTestRules(x402GatePublicRule()), nil).
				Times(1)
			suite.mockPaymentFacilitatorClient.EXPECT().Get(gomock.Any()).Return(x402GateFacilitators(), nil).Times(1)

			suite.mockX402PaymentVerifier.EXPECT().
				RecoverPayer(gomock.Any(), gomock.Any()).
				Return(common.HexToAddress(x402TestPayerHex), nil).
				Times(1)

			suite.mockX402FacilitatorClient.EXPECT().
				Verify(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
				Return(nil, errors.New("connection refused")).
				Times(1)
		}

		t.Run("Should return 502", func(t *testing.T) {
			t.Parallel()

			suite := WhenHandlingX402GateBeforeEach(t)
			initSuite(suite)

			header := x402GatePaymentHeader(t, "eip155:80002", x402TestPayToHex, "10000")
			resp, err := suite.app.Test(x402GateRequest(header))

			assert.NoError(t, err)
			assert.Equal(t, fiber.StatusBadGateway, resp.StatusCode)
		})
	})

	t.Run("Given the payment is valid and settlement succeeds", func(t *testing.T) {
		t.Parallel()

		initSuite := func(suite *WhenHandlingX402GateTestingSuite) {
			suite.mockPaymentRuleClient.EXPECT().
				List(gomock.Any()).
				Return(x402GateTestRules(x402GatePublicRule()), nil).
				Times(1)
			suite.mockPaymentFacilitatorClient.EXPECT().Get(gomock.Any()).Return(x402GateFacilitators(), nil).Times(1)

			suite.mockX402PaymentVerifier.EXPECT().
				RecoverPayer(gomock.Any(), gomock.Any()).
				Return(common.HexToAddress(x402TestPayerHex), nil).
				Times(1)

			suite.mockX402FacilitatorClient.EXPECT().
				Verify(gomock.Any(), "https://x402-amoy.polygon.technology", gomock.Any(), gomock.Any()).
				Return(&x402.VerifyResponse{IsValid: true, Payer: x402TestPayerHex}, nil).
				Times(1)

			suite.mockX402FacilitatorClient.EXPECT().
				Settle(gomock.Any(), "https://x402-amoy.polygon.technology", gomock.Any(), gomock.Any()).
				Return(&x402.SettleResponse{Success: true, Transaction: "0xtxhash", Network: "eip155:80002", Payer: x402TestPayerHex}, nil).
				Times(1)
		}

		t.Run("Should serve the response with a PAYMENT-RESPONSE header", func(t *testing.T) {
			t.Parallel()

			suite := WhenHandlingX402GateBeforeEach(t)
			initSuite(suite)

			header := x402GatePaymentHeader(t, "eip155:80002", x402TestPayToHex, "10000")
			resp, err := suite.app.Test(x402GateRequest(header))

			assert.NoError(t, err)
			assert.Equal(t, fiber.StatusOK, resp.StatusCode)
			assert.True(t, suite.handlerCalled)
			assert.Equal(t, common.HexToAddress(x402TestPayerHex).Hex(), suite.seenPayer)

			raw := resp.Header.Get(middlewares.PaymentResponseHeader)
			assert.NotEmpty(t, raw)
			decoded, err := base64.StdEncoding.DecodeString(raw)
			assert.NoError(t, err)
			var settle x402.SettleResponse
			assert.NoError(t, json.Unmarshal(decoded, &settle))
			assert.True(t, settle.Success)
			assert.Equal(t, "0xtxhash", settle.Transaction)
		})
	})

	t.Run("Given the handler fails after a verified payment", func(t *testing.T) {
		t.Parallel()

		initSuite := func(suite *WhenHandlingX402GateTestingSuite) {
			suite.mockPaymentRuleClient.EXPECT().
				List(gomock.Any()).
				Return(x402GateTestRules(x402GatePublicRule()), nil).
				Times(1)
			suite.mockPaymentFacilitatorClient.EXPECT().Get(gomock.Any()).Return(x402GateFacilitators(), nil).Times(1)

			suite.mockX402PaymentVerifier.EXPECT().
				RecoverPayer(gomock.Any(), gomock.Any()).
				Return(common.HexToAddress(x402TestPayerHex), nil).
				Times(1)

			suite.mockX402FacilitatorClient.EXPECT().
				Verify(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
				Return(&x402.VerifyResponse{IsValid: true}, nil).
				Times(1)
		}

		t.Run("Should not settle the payment", func(t *testing.T) {
			t.Parallel()

			suite := WhenHandlingX402GateBeforeEach(t)
			initSuite(suite)

			failingApp := fiber.New()
			failingApp.Post("/responses", suite.sut.Handle, func(ctx *fiber.Ctx) error {
				return ctx.Status(fiber.StatusBadGateway).JSON(fiber.Map{"code": "runtime_failed"})
			})

			header := x402GatePaymentHeader(t, "eip155:80002", x402TestPayToHex, "10000")
			resp, err := failingApp.Test(x402GateRequest(header))

			assert.NoError(t, err)
			assert.Equal(t, fiber.StatusBadGateway, resp.StatusCode)
		})
	})

	t.Run("Given settlement fails after the handler ran", func(t *testing.T) {
		t.Parallel()

		initSuite := func(suite *WhenHandlingX402GateTestingSuite) {
			suite.mockPaymentRuleClient.EXPECT().
				List(gomock.Any()).
				Return(x402GateTestRules(x402GatePublicRule()), nil).
				Times(1)
			suite.mockPaymentFacilitatorClient.EXPECT().Get(gomock.Any()).Return(x402GateFacilitators(), nil).Times(1)

			suite.mockX402PaymentVerifier.EXPECT().
				RecoverPayer(gomock.Any(), gomock.Any()).
				Return(common.HexToAddress(x402TestPayerHex), nil).
				Times(1)

			suite.mockX402FacilitatorClient.EXPECT().
				Verify(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
				Return(&x402.VerifyResponse{IsValid: true}, nil).
				Times(1)

			suite.mockX402FacilitatorClient.EXPECT().
				Settle(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
				Return(&x402.SettleResponse{Success: false, ErrorReason: "nonce_already_used"}, nil).
				Times(1)
		}

		t.Run("Should discard the response body and return 402", func(t *testing.T) {
			t.Parallel()

			suite := WhenHandlingX402GateBeforeEach(t)
			initSuite(suite)

			header := x402GatePaymentHeader(t, "eip155:80002", x402TestPayToHex, "10000")
			resp, err := suite.app.Test(x402GateRequest(header))

			assert.NoError(t, err)
			assert.Equal(t, fiber.StatusPaymentRequired, resp.StatusCode)
			assert.True(t, suite.handlerCalled)

			bodyBytes, _ := io.ReadAll(resp.Body)
			assert.NotContains(t, string(bodyBytes), "42")
			assert.Equal(t, "settlement_failed", decodePaymentRequiredHeader(t, resp).Error)
			assert.Empty(t, resp.Header.Get(middlewares.PaymentResponseHeader))
		})
	})

	t.Run("Given a JWT whose address has a zero-price rule", func(t *testing.T) {
		t.Parallel()

		initSuite := func(suite *WhenHandlingX402GateTestingSuite) {
			suite.mockPaymentRuleClient.EXPECT().
				List(gomock.Any()).
				Return(x402GateTestRules(
					x402GateFreeRule(strings.ToLower(x402TestPayerHex)),
				), nil).
				Times(1)
			suite.mockJwtValidator.EXPECT().
				Verify("valid.jwt.token").
				Return(common.HexToAddress(x402TestPayerHex), nil).
				Times(1)
		}

		t.Run("Should serve free without a signature or the facilitator", func(t *testing.T) {
			t.Parallel()

			suite := WhenHandlingX402GateBeforeEach(t)
			initSuite(suite)

			resp, err := suite.app.Test(x402GateBearerRequest("valid.jwt.token", ""))

			assert.NoError(t, err)
			assert.Equal(t, fiber.StatusOK, resp.StatusCode)
			assert.True(t, suite.handlerCalled)
			// Identity comes from the JWT, not a recovered signature.
			assert.Equal(t, common.HexToAddress(x402TestPayerHex).Hex(), suite.seenAddress)
			assert.Empty(t, suite.seenPayer)
		})
	})

	t.Run("Given a JWT whose address has only a paid rule", func(t *testing.T) {
		t.Parallel()

		initSuite := func(suite *WhenHandlingX402GateTestingSuite) {
			suite.mockPaymentRuleClient.EXPECT().
				List(gomock.Any()).
				Return(x402GateTestRules(x402GatePublicRule()), nil).
				Times(1)
			suite.mockJwtValidator.EXPECT().
				Verify("valid.jwt.token").
				Return(common.HexToAddress(x402TestPayerHex), nil).
				Times(1)
			suite.mockPaymentFacilitatorClient.EXPECT().Get(gomock.Any()).Return(x402GateFacilitators(), nil).Times(1)
		}

		t.Run("Should fall through to the 402 so the caller signs", func(t *testing.T) {
			t.Parallel()

			suite := WhenHandlingX402GateBeforeEach(t)
			initSuite(suite)

			resp, err := suite.app.Test(x402GateBearerRequest("valid.jwt.token", ""))

			assert.NoError(t, err)
			assert.Equal(t, fiber.StatusPaymentRequired, resp.StatusCode)
			assert.False(t, suite.handlerCalled)
		})
	})

	t.Run("Given an invalid JWT alongside a payment signature", func(t *testing.T) {
		t.Parallel()

		initSuite := func(suite *WhenHandlingX402GateTestingSuite) {
			suite.mockPaymentRuleClient.EXPECT().
				List(gomock.Any()).
				Return(x402GateTestRules(
					x402GateFreeRule(strings.ToLower(x402TestPayerHex)),
				), nil).
				Times(1)
			suite.mockJwtValidator.EXPECT().
				Verify("expired.jwt.token").
				Return(common.Address{}, errors.New("token expired")).
				Times(1)
			suite.mockPaymentFacilitatorClient.EXPECT().Get(gomock.Any()).Return(x402GateFacilitators(), nil).Times(1)
			suite.mockX402PaymentVerifier.EXPECT().
				RecoverPayer(gomock.Any(), gomock.Any()).
				Return(common.HexToAddress(x402TestPayerHex), nil).
				Times(1)
		}

		t.Run("Should ignore the JWT and honor the signature path", func(t *testing.T) {
			t.Parallel()

			suite := WhenHandlingX402GateBeforeEach(t)
			initSuite(suite)

			header := x402GatePaymentHeader(t, "eip155:80002", x402TestAgentWalletHex, "0")
			resp, err := suite.app.Test(x402GateBearerRequest("expired.jwt.token", header))

			assert.NoError(t, err)
			assert.Equal(t, fiber.StatusOK, resp.StatusCode)
			assert.True(t, suite.handlerCalled)
		})
	})
}
