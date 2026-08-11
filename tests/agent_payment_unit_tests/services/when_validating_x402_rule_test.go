package services_test

import (
	"context"
	"testing"

	"github.com/6022-labs/agentic-agent-node/src/agent_payment/models"
	"github.com/6022-labs/agentic-agent-node/src/agent_payment/models/enums"
	"github.com/6022-labs/agentic-agent-node/src/agent_payment/services"
	apperrors "github.com/6022-labs/agentic-agent-node/src/common/errors"
	runtime_common_settings "github.com/6022-labs/agentic-agent-node/src/runtime_common/settings"
	payment_repositories_mocks "github.com/6022-labs/agentic-agent-node/tests/agent_payment_mocks/repositories_mocks"
	payment_interfaces_mocks "github.com/6022-labs/agentic-agent-node/tests/agent_payment_mocks/services_mocks/interfaces_mocks"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

type WhenValidatingX402RuleTestingSuite struct {
	sut *services.X402PaymentMethodStrategy

	mockErc20Validator *payment_interfaces_mocks.MockErc20ValidatorInterface
}

func WhenValidatingX402RuleBeforeEach(t *testing.T) *WhenValidatingX402RuleTestingSuite {
	return validatingX402RuleSuite(t, &runtime_common_settings.PaymentFacilitatorsSettings{
		"eip155:80002": {BaseUrl: "https://x402-amoy.polygon.technology"},
		"eip155:84532": {BaseUrl: "https://x402.org/facilitator"},
	})
}

// validatingX402RuleSuite builds the suite with a given facilitator config so the
// facilitator-gate cases can run with no facilitator configured. Chains 80002
// and 84532 are reachable; any other chain (e.g. eip155:1) is not.
func validatingX402RuleSuite(
	t *testing.T,
	facilitators *runtime_common_settings.PaymentFacilitatorsSettings,
) *WhenValidatingX402RuleTestingSuite {
	mockController := gomock.NewController(t)

	mockErc20Validator := payment_interfaces_mocks.NewMockErc20ValidatorInterface(mockController)
	mockConfigRepository := payment_repositories_mocks.NewMockPaymentRuleX402ConfigRepositoryInterface(mockController)
	evmChainsSettings := &runtime_common_settings.EvmChainsSettings{
		"80002": {ChainId: "80002", HttpUrl: "http://localhost:8545"},
		"84532": {ChainId: "84532", HttpUrl: "http://localhost:8546"},
	}

	return &WhenValidatingX402RuleTestingSuite{
		sut: services.NewX402PaymentMethodStrategy(
			facilitators,
			evmChainsSettings,
			mockErc20Validator,
			mockConfigRepository,
		),
		mockErc20Validator: mockErc20Validator,
	}
}

const validConfig = `{"network":"eip155:80002","asset":"0x41E94Eb019C0762f9Bfcf9Fb1E58725BfB0e7582","payTo":"0xf39Fd6e51aad88F6F4ce6aB8827279cffFb92266","addressPatterns":[".*"," 0x70997970c51812dc3a010c7d01b50e0d17dc79c8 "]}`

func TestWhenValidatingX402Rule(t *testing.T) {
	t.Parallel()

	t.Run("Given the config is valid", func(t *testing.T) {
		t.Parallel()

		initSuite := func(suite *WhenValidatingX402RuleTestingSuite) {
			// Distinct from the seed defaults to prove the derived domain wins.
			suite.mockErc20Validator.EXPECT().
				ValidateErc20(gomock.Any(), "80002", "0x41E94Eb019C0762f9Bfcf9Fb1E58725BfB0e7582").
				Return(enums.AssetTransferMethodEip3009, "USD Coin", "7", nil).
				Times(1)
		}

		t.Run("Should sanitize and set the derived EIP-712 domain", func(t *testing.T) {
			t.Parallel()

			suite := WhenValidatingX402RuleBeforeEach(t)
			initSuite(suite)

			request := generateSaveRequest(validConfig, "10000")
			err := suite.sut.Validate(context.Background(), request)

			assert.NoError(t, err)
			config := &models.PaymentRuleX402Config{}
			assert.NoError(t, config.Parse(request.Config))
			assert.Equal(t, ".*", config.AddressPatterns[0])
			assert.Equal(t, "0x70997970c51812dc3a010c7d01b50e0d17dc79c8", config.AddressPatterns[1])
			assert.Equal(t, enums.AssetTransferMethodEip3009, config.AssetTransferMethod)
			assert.Equal(t, "USD Coin", *config.Eip712Name)
			assert.Equal(t, "7", *config.Eip712Version)
		})
	})

	t.Run("Given the token has no usable EIP-712 domain", func(t *testing.T) {
		t.Parallel()

		initSuite := func(suite *WhenValidatingX402RuleTestingSuite) {
			suite.mockErc20Validator.EXPECT().
				ValidateErc20(gomock.Any(), "80002", "0x41E94Eb019C0762f9Bfcf9Fb1E58725BfB0e7582").
				Return(enums.AssetTransferMethodPermit2, "", "", nil).
				Times(1)
		}

		t.Run("Should set permit2 with an empty domain", func(t *testing.T) {
			t.Parallel()

			suite := WhenValidatingX402RuleBeforeEach(t)
			initSuite(suite)

			request := generateSaveRequest(validConfig, "10000")
			err := suite.sut.Validate(context.Background(), request)

			assert.NoError(t, err)
			config := &models.PaymentRuleX402Config{}
			assert.NoError(t, config.Parse(request.Config))
			assert.Equal(t, enums.AssetTransferMethodPermit2, config.AssetTransferMethod)
			assert.Nil(t, config.Eip712Name)
			assert.Nil(t, config.Eip712Version)
		})
	})

	t.Run("Given the network is not eip155", func(t *testing.T) {
		t.Parallel()

		t.Run("Should return a validation error on config.network", func(t *testing.T) {
			t.Parallel()

			suite := WhenValidatingX402RuleBeforeEach(t)

			request := generateSaveRequest(
				`{"network":"solana:mainnet","asset":"0x41E94Eb019C0762f9Bfcf9Fb1E58725BfB0e7582","payTo":"0xf39Fd6e51aad88F6F4ce6aB8827279cffFb92266","addressPatterns":[".*"]}`,
				"10000",
			)
			err := suite.sut.Validate(context.Background(), request)

			var validationErr *apperrors.ValidationError
			assert.ErrorAs(t, err, &validationErr)
			assert.Equal(t, "config.network", validationErr.Field)
		})
	})

	t.Run("Given the eip155 chain id is not a positive integer", func(t *testing.T) {
		t.Parallel()

		for _, network := range []string{"eip155:abc", "eip155:-5"} {
			t.Run("Should return a validation error on config.network for "+network, func(t *testing.T) {
				t.Parallel()

				suite := WhenValidatingX402RuleBeforeEach(t)

				request := generateSaveRequest(
					`{"network":"`+network+`","asset":"0x41E94Eb019C0762f9Bfcf9Fb1E58725BfB0e7582","payTo":"0xf39Fd6e51aad88F6F4ce6aB8827279cffFb92266","addressPatterns":[".*"]}`,
					"10000",
				)
				err := suite.sut.Validate(context.Background(), request)

				var validationErr *apperrors.ValidationError
				assert.ErrorAs(t, err, &validationErr)
				assert.Equal(t, "config.network", validationErr.Field)
			})
		}
	})

	t.Run("Given an unreachable chain", func(t *testing.T) {
		t.Parallel()

		t.Run("Should return a validation error on config.network", func(t *testing.T) {
			t.Parallel()

			suite := WhenValidatingX402RuleBeforeEach(t)

			request := generateSaveRequest(
				`{"network":"eip155:1","asset":"0x41E94Eb019C0762f9Bfcf9Fb1E58725BfB0e7582","payTo":"0xf39Fd6e51aad88F6F4ce6aB8827279cffFb92266","addressPatterns":[".*"]}`,
				"10000",
			)
			err := suite.sut.Validate(context.Background(), request)

			var validationErr *apperrors.ValidationError
			assert.ErrorAs(t, err, &validationErr)
			assert.Equal(t, "config.network", validationErr.Field)
		})
	})

	t.Run("Given a paid rule on a reachable chain other than the identity chain", func(t *testing.T) {
		t.Parallel()

		t.Run("Should be allowed — payments are decoupled from identity", func(t *testing.T) {
			t.Parallel()

			suite := WhenValidatingX402RuleBeforeEach(t)
			suite.mockErc20Validator.EXPECT().
				ValidateErc20(gomock.Any(), "84532", "0x41E94Eb019C0762f9Bfcf9Fb1E58725BfB0e7582").
				Return(enums.AssetTransferMethodEip3009, "USD Coin", "2", nil).
				Times(1)

			request := generateSaveRequest(
				`{"network":"eip155:84532","asset":"0x41E94Eb019C0762f9Bfcf9Fb1E58725BfB0e7582","payTo":"0xf39Fd6e51aad88F6F4ce6aB8827279cffFb92266","addressPatterns":[".*"]}`,
				"10000",
			)
			err := suite.sut.Validate(context.Background(), request)

			assert.NoError(t, err)
		})
	})

	t.Run("Given the asset is not a hex address", func(t *testing.T) {
		t.Parallel()

		t.Run("Should return a validation error on config.asset", func(t *testing.T) {
			t.Parallel()

			suite := WhenValidatingX402RuleBeforeEach(t)

			request := generateSaveRequest(
				`{"network":"eip155:80002","asset":"not-hex","payTo":"0xf39Fd6e51aad88F6F4ce6aB8827279cffFb92266","addressPatterns":[".*"]}`,
				"10000",
			)
			err := suite.sut.Validate(context.Background(), request)

			var validationErr *apperrors.ValidationError
			assert.ErrorAs(t, err, &validationErr)
			assert.Equal(t, "config.asset", validationErr.Field)
		})
	})

	t.Run("Given the payTo is not a hex address", func(t *testing.T) {
		t.Parallel()

		t.Run("Should return a validation error on config.payTo", func(t *testing.T) {
			t.Parallel()

			suite := WhenValidatingX402RuleBeforeEach(t)

			request := generateSaveRequest(
				`{"network":"eip155:80002","asset":"0x41E94Eb019C0762f9Bfcf9Fb1E58725BfB0e7582","payTo":"not-hex","addressPatterns":[".*"]}`,
				"10000",
			)
			err := suite.sut.Validate(context.Background(), request)

			var validationErr *apperrors.ValidationError
			assert.ErrorAs(t, err, &validationErr)
			assert.Equal(t, "config.payTo", validationErr.Field)
		})
	})

	t.Run("Given an empty address pattern", func(t *testing.T) {
		t.Parallel()

		t.Run("Should return a validation error on config.addressPatterns", func(t *testing.T) {
			t.Parallel()

			suite := WhenValidatingX402RuleBeforeEach(t)

			request := generateSaveRequest(
				`{"network":"eip155:80002","asset":"0x41E94Eb019C0762f9Bfcf9Fb1E58725BfB0e7582","payTo":"0xf39Fd6e51aad88F6F4ce6aB8827279cffFb92266","addressPatterns":["  "]}`,
				"10000",
			)
			err := suite.sut.Validate(context.Background(), request)

			var validationErr *apperrors.ValidationError
			assert.ErrorAs(t, err, &validationErr)
			assert.Equal(t, "config.addressPatterns", validationErr.Field)
		})
	})

	t.Run("Given no address patterns", func(t *testing.T) {
		t.Parallel()

		t.Run("Should return a validation error on config.addressPatterns", func(t *testing.T) {
			t.Parallel()

			suite := WhenValidatingX402RuleBeforeEach(t)

			request := generateSaveRequest(
				`{"network":"eip155:80002","asset":"0x41E94Eb019C0762f9Bfcf9Fb1E58725BfB0e7582","payTo":"0xf39Fd6e51aad88F6F4ce6aB8827279cffFb92266","addressPatterns":[]}`,
				"10000",
			)
			err := suite.sut.Validate(context.Background(), request)

			var validationErr *apperrors.ValidationError
			assert.ErrorAs(t, err, &validationErr)
			assert.Equal(t, "config.addressPatterns", validationErr.Field)
		})
	})

	t.Run("Given an invalid regex pattern", func(t *testing.T) {
		t.Parallel()

		t.Run("Should return a validation error on config.addressPatterns", func(t *testing.T) {
			t.Parallel()

			suite := WhenValidatingX402RuleBeforeEach(t)

			request := generateSaveRequest(
				`{"network":"eip155:80002","asset":"0x41E94Eb019C0762f9Bfcf9Fb1E58725BfB0e7582","payTo":"0xf39Fd6e51aad88F6F4ce6aB8827279cffFb92266","addressPatterns":["[invalid"]}`,
				"10000",
			)
			err := suite.sut.Validate(context.Background(), request)

			var validationErr *apperrors.ValidationError
			assert.ErrorAs(t, err, &validationErr)
			assert.Equal(t, "config.addressPatterns", validationErr.Field)
		})
	})

	t.Run("Given a paid rule and the node chain has no facilitator", func(t *testing.T) {
		t.Parallel()

		initSuite := func(suite *WhenValidatingX402RuleTestingSuite) {
			suite.mockErc20Validator.EXPECT().
				ValidateErc20(gomock.Any(), gomock.Any(), gomock.Any()).
				Return(enums.AssetTransferMethodEip3009, "USDC", "2", nil).
				Times(1)
		}

		t.Run("Should return a validation error on config.network", func(t *testing.T) {
			t.Parallel()

			suite := validatingX402RuleSuite(t, &runtime_common_settings.PaymentFacilitatorsSettings{})
			initSuite(suite)

			request := generateSaveRequest(validConfig, "10000")
			err := suite.sut.Validate(context.Background(), request)

			var validationErr *apperrors.ValidationError
			assert.ErrorAs(t, err, &validationErr)
			assert.Equal(t, "config.network", validationErr.Field)
		})
	})

	t.Run("Given a free rule and the node chain has no facilitator", func(t *testing.T) {
		t.Parallel()

		initSuite := func(suite *WhenValidatingX402RuleTestingSuite) {
			suite.mockErc20Validator.EXPECT().
				ValidateErc20(gomock.Any(), gomock.Any(), gomock.Any()).
				Return(enums.AssetTransferMethodEip3009, "USDC", "2", nil).
				Times(1)
		}

		t.Run("Should be allowed (free rules never settle)", func(t *testing.T) {
			t.Parallel()

			suite := validatingX402RuleSuite(t, &runtime_common_settings.PaymentFacilitatorsSettings{})
			initSuite(suite)

			request := generateSaveRequest(validConfig, "0")
			err := suite.sut.Validate(context.Background(), request)

			assert.NoError(t, err)
		})
	})

	t.Run("Given the asset is not an ERC-20 contract", func(t *testing.T) {
		t.Parallel()

		initSuite := func(suite *WhenValidatingX402RuleTestingSuite) {
			suite.mockErc20Validator.EXPECT().
				ValidateErc20(gomock.Any(), gomock.Any(), gomock.Any()).
				Return(enums.AssetTransferMethod(""), "", "", assert.AnError).
				Times(1)
		}

		t.Run("Should return a validation error on config.asset", func(t *testing.T) {
			t.Parallel()

			suite := WhenValidatingX402RuleBeforeEach(t)
			initSuite(suite)

			request := generateSaveRequest(validConfig, "10000")
			err := suite.sut.Validate(context.Background(), request)

			var validationErr *apperrors.ValidationError
			assert.ErrorAs(t, err, &validationErr)
			assert.Equal(t, "config.asset", validationErr.Field)
		})
	})
}
