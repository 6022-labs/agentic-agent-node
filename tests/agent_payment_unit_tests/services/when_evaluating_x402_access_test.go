package services_test

import (
	"context"
	"crypto/ecdsa"
	"encoding/json"
	"strings"
	"testing"

	"github.com/6022-labs/agentic-agent-node/src/agent_payment/models"
	"github.com/6022-labs/agentic-agent-node/src/agent_payment/models/enums"
	"github.com/6022-labs/agentic-agent-node/src/agent_payment/services"
	payment_interfaces "github.com/6022-labs/agentic-agent-node/src/agent_payment/services/interfaces"
	apperrors "github.com/6022-labs/agentic-agent-node/src/common/errors"
	"github.com/6022-labs/agentic-agent-node/src/common/x402"
	runtime_common_settings "github.com/6022-labs/agentic-agent-node/src/runtime_common/settings"
	payment_repositories_mocks "github.com/6022-labs/agentic-agent-node/tests/agent_payment_mocks/repositories_mocks"
	payment_interfaces_mocks "github.com/6022-labs/agentic-agent-node/tests/agent_payment_mocks/services_mocks/interfaces_mocks"
	"github.com/6022-labs/agentic-agent-node/tests/common_tests/model_builders"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

type WhenEvaluatingX402AccessTestingSuite struct {
	sut *services.X402PaymentReceivePolicyEvaluator

	mockPaymentReceivePolicyProvider           *payment_interfaces_mocks.MockPaymentReceivePolicyProviderInterface
	mockPaymentReceiveRuleRepository           *payment_repositories_mocks.MockPaymentReceiveRuleRepositoryInterface
	mockPaymentReceiveRuleX402ConfigRepository *payment_repositories_mocks.MockPaymentReceiveRuleX402ConfigRepositoryInterface
	callerKey                                  *ecdsa.PrivateKey
}

func WhenEvaluatingX402AccessBeforeEach(t *testing.T) *WhenEvaluatingX402AccessTestingSuite {
	mockController := gomock.NewController(t)

	callerKey, err := crypto.GenerateKey()
	assert.NoError(t, err)

	mockPaymentReceivePolicyProvider := payment_interfaces_mocks.NewMockPaymentReceivePolicyProviderInterface(
		mockController,
	)
	mockPaymentReceiveRuleRepository := payment_repositories_mocks.NewMockPaymentReceiveRuleRepositoryInterface(
		mockController,
	)
	mockPaymentReceiveRuleX402ConfigRepository :=
		payment_repositories_mocks.NewMockPaymentReceiveRuleX402ConfigRepositoryInterface(mockController)

	// The node can settle on the rules' network; an unsettleable rule overrides this per Given.
	facilitators := &runtime_common_settings.PaymentReceiveSettings{
		"eip155:80002": runtime_common_settings.PaymentFacilitator{
			Network:           "eip155:80002",
			BaseUrl:           "https://facilitator.test",
			MaxTimeoutSeconds: 90,
		},
	}

	return &WhenEvaluatingX402AccessTestingSuite{
		sut: services.NewX402PaymentReceivePolicyEvaluator(
			mockPaymentReceivePolicyProvider,
			mockPaymentReceiveRuleRepository,
			mockPaymentReceiveRuleX402ConfigRepository,
			services.NewX402PaymentVerifier(),
			facilitators,
		),

		mockPaymentReceivePolicyProvider:           mockPaymentReceivePolicyProvider,
		mockPaymentReceiveRuleRepository:           mockPaymentReceiveRuleRepository,
		mockPaymentReceiveRuleX402ConfigRepository: mockPaymentReceiveRuleX402ConfigRepository,
		callerKey: callerKey,
	}
}

// expectRules stubs the policy plus one rule per (price, config) pair.
func (s *WhenEvaluatingX402AccessTestingSuite) expectRules(
	prices []string,
	configs []*models.PaymentReceiveRuleX402Config,
) {
	agent := model_builders.NewAgentBuilder().Build()
	policy := model_builders.NewPaymentReceivePolicyBuilder().WithAgentId(agent.Id).Build()

	s.mockPaymentReceivePolicyProvider.EXPECT().Ensure(gomock.Any()).Return(policy, agent, nil).Times(1)

	rules := make([]models.PaymentReceiveRule, 0, len(prices))
	for i := range prices {
		rule := model_builders.NewPaymentReceiveRuleBuilder().WithPolicyId(policy.Id).Build()
		rule.Price = prices[i]
		rule.Method = enums.PaymentMethodX402
		rules = append(rules, *rule)

		config := configs[i]
		config.PaymentReceiveRuleId = rule.Id
		s.mockPaymentReceiveRuleX402ConfigRepository.EXPECT().GetByRuleId(gomock.Any(), rule.Id).Return(config, nil)
	}
	s.mockPaymentReceiveRuleRepository.EXPECT().ListByPolicyId(gomock.Any(), policy.Id).Return(rules, nil).Times(1)
}

// signedHeader signs the quote the way an x402 client would and encodes the header.
func (s *WhenEvaluatingX402AccessTestingSuite) signedHeader(
	t *testing.T,
	config *models.PaymentReceiveRuleX402Config,
) string {
	t.Helper()

	from := crypto.PubkeyToAddress(s.callerKey.PublicKey)
	payload := signX402Payment(t, s.callerKey, x402TestAuthorization(from.Hex(), config), config)
	header, err := x402.EncodeBase64Json(payload)
	assert.NoError(t, err)
	return header
}

func publicConfig() *models.PaymentReceiveRuleX402Config {
	config := x402TestAcceptOption()
	config.AddressPatterns = []string{".*"}
	return config
}

func scopedConfig(patterns ...string) *models.PaymentReceiveRuleX402Config {
	config := x402TestAcceptOption()
	config.AddressPatterns = patterns
	return config
}

func TestWhenEvaluatingX402Access(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	t.Run("Given a JWT-proven caller a free rule covers", func(t *testing.T) {
		t.Parallel()

		t.Run("Should grant free access without a signature", func(t *testing.T) {
			t.Parallel()

			suite := WhenEvaluatingX402AccessBeforeEach(t)
			caller := crypto.PubkeyToAddress(suite.callerKey.PublicKey)
			suite.expectRules(
				[]string{"0"},
				[]*models.PaymentReceiveRuleX402Config{scopedConfig(strings.ToLower(caller.Hex()))},
			)

			decision, err := suite.sut.Evaluate(ctx, &payment_interfaces.PaymentReceiveAccessCredentials{
				Address: caller.Hex(),
			})

			assert.NoError(t, err)
			assert.Equal(t, enums.PaymentReceiveAccessOutcomeFree, decision.Outcome)
			assert.Equal(t, caller.Hex(), decision.Payer)
		})
	})

	t.Run("Given a JWT-proven caller no free rule covers", func(t *testing.T) {
		t.Parallel()

		t.Run("Should require payment", func(t *testing.T) {
			t.Parallel()

			suite := WhenEvaluatingX402AccessBeforeEach(t)
			suite.expectRules([]string{"0"}, []*models.PaymentReceiveRuleX402Config{
				scopedConfig("0x0000000000000000000000000000000000000000"),
			})

			decision, err := suite.sut.Evaluate(ctx, &payment_interfaces.PaymentReceiveAccessCredentials{
				Address: crypto.PubkeyToAddress(suite.callerKey.PublicKey).Hex(),
			})

			assert.NoError(t, err)
			assert.Equal(t, enums.PaymentReceiveAccessOutcomePaymentRequired, decision.Outcome)
		})
	})

	t.Run("Given no credentials at all", func(t *testing.T) {
		t.Parallel()

		t.Run("Should advertise only the public rules", func(t *testing.T) {
			t.Parallel()

			suite := WhenEvaluatingX402AccessBeforeEach(t)
			scoped := scopedConfig(strings.ToLower("0xf39Fd6e51aad88F6F4ce6aB8827279cffFb92266"))
			// Distinct payTo so the advertised entry identifies which rule was chosen.
			scoped.PayTo = "0x000000000000000000000000000000000000dEaD"
			suite.expectRules(
				[]string{"10000", "10000"},
				[]*models.PaymentReceiveRuleX402Config{publicConfig(), scoped},
			)

			decision, err := suite.sut.Evaluate(ctx, &payment_interfaces.PaymentReceiveAccessCredentials{})

			assert.NoError(t, err)
			assert.Equal(t, enums.PaymentReceiveAccessOutcomePaymentRequired, decision.Outcome)
			// The scoped rule stays hidden: advertising it would leak its address patterns.
			assert.Len(t, decision.Accepts, 1)
			assert.Equal(t, publicConfig().PayTo, decision.Accepts[0].PayTo)
			assert.Equal(t, 90, decision.Accepts[0].MaxTimeoutSeconds)
			// eip3009 needs the token's domain to sign; permit2 rules carry none.
			assert.Equal(t, "eip3009", decision.Accepts[0].Extra["assetTransferMethod"])
		})
	})

	t.Run("Given a paid rule on a network with no facilitator", func(t *testing.T) {
		t.Parallel()

		t.Run("Should not advertise it, since it could never settle", func(t *testing.T) {
			t.Parallel()

			suite := WhenEvaluatingX402AccessBeforeEach(t)
			unsettleable := publicConfig()
			unsettleable.Network = "eip155:1"
			suite.expectRules([]string{"10000"}, []*models.PaymentReceiveRuleX402Config{unsettleable})

			decision, err := suite.sut.Evaluate(ctx, &payment_interfaces.PaymentReceiveAccessCredentials{})

			assert.NoError(t, err)
			assert.Empty(t, decision.Accepts)
		})
	})

	t.Run("Given a malformed payment signature", func(t *testing.T) {
		t.Parallel()

		t.Run("Should say so in the refusal", func(t *testing.T) {
			t.Parallel()

			suite := WhenEvaluatingX402AccessBeforeEach(t)
			suite.expectRules([]string{"10000"}, []*models.PaymentReceiveRuleX402Config{publicConfig()})

			decision, err := suite.sut.Evaluate(ctx, &payment_interfaces.PaymentReceiveAccessCredentials{
				PaymentSignature: "not-base64-json",
			})

			assert.NoError(t, err)
			assert.Equal(t, enums.PaymentReceiveAccessOutcomePaymentRequired, decision.Outcome)
			assert.Equal(t, "malformed_payment_signature", decision.ErrorCode)
		})
	})

	t.Run("Given a signature paying a payTo no rule uses", func(t *testing.T) {
		t.Parallel()

		t.Run("Should refuse as unmatched", func(t *testing.T) {
			t.Parallel()

			suite := WhenEvaluatingX402AccessBeforeEach(t)
			signedAgainst := publicConfig()
			signedAgainst.PayTo = "0x000000000000000000000000000000000000dEaD"
			header := suite.signedHeader(t, signedAgainst)

			suite.expectRules([]string{"10000"}, []*models.PaymentReceiveRuleX402Config{publicConfig()})

			decision, err := suite.sut.Evaluate(ctx, &payment_interfaces.PaymentReceiveAccessCredentials{
				PaymentSignature: header,
			})

			assert.NoError(t, err)
			assert.Equal(t, "no_matching_payment_requirements", decision.ErrorCode)
		})
	})

	t.Run("Given a signature that covers a public paid rule", func(t *testing.T) {
		t.Parallel()

		t.Run("Should be payable and carry what settles it", func(t *testing.T) {
			t.Parallel()

			suite := WhenEvaluatingX402AccessBeforeEach(t)
			config := publicConfig()
			header := suite.signedHeader(t, config)
			suite.expectRules([]string{testPrice}, []*models.PaymentReceiveRuleX402Config{config})

			decision, err := suite.sut.Evaluate(ctx, &payment_interfaces.PaymentReceiveAccessCredentials{
				PaymentSignature: header,
			})

			assert.NoError(t, err)
			assert.Equal(t, enums.PaymentReceiveAccessOutcomePayable, decision.Outcome)
			assert.Equal(t, crypto.PubkeyToAddress(suite.callerKey.PublicKey).Hex(), decision.Payer)
			assert.NotNil(t, decision.Settlement)
			assert.Equal(t, "https://facilitator.test", decision.Settlement.FacilitatorBaseUrl)
			assert.Equal(t, testPrice, decision.Settlement.Requirements.Amount)
		})
	})

	t.Run("Given the signed amount is below the rule's price", func(t *testing.T) {
		t.Parallel()

		t.Run("Should refuse as insufficient", func(t *testing.T) {
			t.Parallel()

			suite := WhenEvaluatingX402AccessBeforeEach(t)
			config := publicConfig()
			header := suite.signedHeader(t, config)
			// The signature is worth testPrice; the rule now asks for more.
			suite.expectRules([]string{"999999999"}, []*models.PaymentReceiveRuleX402Config{config})

			decision, err := suite.sut.Evaluate(ctx, &payment_interfaces.PaymentReceiveAccessCredentials{
				PaymentSignature: header,
			})

			assert.NoError(t, err)
			assert.Equal(t, "insufficient_amount", decision.ErrorCode)
		})
	})

	t.Run("Given the payer is entitled to a free rule as well", func(t *testing.T) {
		t.Parallel()

		// The cheapest entitled rule wins, so a payer covered by a free rule is never charged.
		t.Run("Should grant free access rather than settle", func(t *testing.T) {
			t.Parallel()

			suite := WhenEvaluatingX402AccessBeforeEach(t)
			caller := crypto.PubkeyToAddress(suite.callerKey.PublicKey)
			paid := publicConfig()
			header := suite.signedHeader(t, paid)
			suite.expectRules(
				[]string{testPrice, "0"},
				[]*models.PaymentReceiveRuleX402Config{paid, scopedConfig(strings.ToLower(caller.Hex()))},
			)

			decision, err := suite.sut.Evaluate(ctx, &payment_interfaces.PaymentReceiveAccessCredentials{
				PaymentSignature: header,
			})

			assert.NoError(t, err)
			assert.Equal(t, enums.PaymentReceiveAccessOutcomeFree, decision.Outcome)
			assert.Nil(t, decision.Settlement)
		})
	})

	t.Run("Given the agent is not minted yet", func(t *testing.T) {
		t.Parallel()

		// Pre-mint there is no policy to materialize, so the route simply has nothing to sell.
		t.Run("Should require payment with nothing to advertise", func(t *testing.T) {
			t.Parallel()

			suite := WhenEvaluatingX402AccessBeforeEach(t)
			suite.mockPaymentReceivePolicyProvider.EXPECT().
				Ensure(gomock.Any()).
				Return(nil, nil, apperrors.NewConflictError("agent_not_minted", "")).
				Times(1)

			decision, err := suite.sut.Evaluate(ctx, &payment_interfaces.PaymentReceiveAccessCredentials{})

			assert.NoError(t, err)
			assert.Equal(t, enums.PaymentReceiveAccessOutcomePaymentRequired, decision.Outcome)
			assert.Empty(t, decision.Accepts)
		})
	})

	t.Run("Given a payload carrying a second, unsigned authorization", func(t *testing.T) {
		t.Parallel()

		// A payload signs exactly one authorization. Attaching an unsigned permit2 blob beside a
		// genuine eip3009 one used to hijack the amount check, since the method-neutral reader
		// takes permit2 first while the eip3009 verifier never looks at it.
		initSuite := func(suite *WhenEvaluatingX402AccessTestingSuite) string {
			config := publicConfig()

			from := crypto.PubkeyToAddress(suite.callerKey.PublicKey)
			auth := x402TestAuthorization(from.Hex(), config)
			auth.Value = "1" // genuinely signed, but worth almost nothing
			payload := signX402Payment(t, suite.callerKey, auth, config)
			payload.Payload.Permit2Authorization = &x402.Permit2Authorization{
				Permitted: x402.Permit2TokenPermissions{Token: config.Asset, Amount: "999999999999"},
				From:      from.Hex(),
				Spender:   x402.Permit2Spender,
				Nonce:     "1",
				Deadline:  "99999999999",
				Witness:   x402.Permit2Witness{To: config.PayTo, ValidAfter: "0"},
			}
			raw, err := json.Marshal(payload)
			assert.NoError(t, err)
			header, err := x402.EncodeBase64Json(json.RawMessage(raw))
			assert.NoError(t, err)

			suite.expectRules([]string{testPrice}, []*models.PaymentReceiveRuleX402Config{config})
			return header
		}

		t.Run("Should not authorize a rule the signed amount does not cover", func(t *testing.T) {
			t.Parallel()

			suite := WhenEvaluatingX402AccessBeforeEach(t)
			header := initSuite(suite)

			decision, err := suite.sut.Evaluate(ctx, &payment_interfaces.PaymentReceiveAccessCredentials{
				PaymentSignature: header,
			})

			assert.NoError(t, err)
			assert.NotEqual(t, enums.PaymentReceiveAccessOutcomePayable, decision.Outcome,
				"paying 1 atomic unit must not authorize a rule priced at %s", testPrice)
		})
	})

	t.Run("Given listing the rules fails", func(t *testing.T) {
		t.Parallel()

		t.Run("Should return an internal error rather than a refusal", func(t *testing.T) {
			t.Parallel()

			suite := WhenEvaluatingX402AccessBeforeEach(t)
			agent := model_builders.NewAgentBuilder().Build()
			policy := model_builders.NewPaymentReceivePolicyBuilder().WithAgentId(agent.Id).Build()
			suite.mockPaymentReceivePolicyProvider.EXPECT().Ensure(gomock.Any()).Return(policy, agent, nil).Times(1)
			suite.mockPaymentReceiveRuleRepository.EXPECT().
				ListByPolicyId(gomock.Any(), policy.Id).
				Return(nil, assert.AnError).
				Times(1)

			decision, err := suite.sut.Evaluate(ctx, &payment_interfaces.PaymentReceiveAccessCredentials{})

			assert.Nil(t, decision)
			var internalError *apperrors.InternalError
			assert.ErrorAs(t, err, &internalError)
		})
	})
}
