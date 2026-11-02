package services_test

import (
	"context"
	"crypto/ecdsa"
	"encoding/json"
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/6022-labs/agentic-agent-node/src/agent_payment/models"
	"github.com/6022-labs/agentic-agent-node/src/agent_payment/models/enums"
	"github.com/6022-labs/agentic-agent-node/src/agent_payment/services"
	payment_interfaces "github.com/6022-labs/agentic-agent-node/src/agent_payment/services/interfaces"
	apperrors "github.com/6022-labs/agentic-agent-node/src/common/errors"
	"github.com/6022-labs/agentic-agent-node/src/common/x402"
	runtime_common_settings "github.com/6022-labs/agentic-agent-node/src/runtime_common/settings"
	payment_repositories_mocks "github.com/6022-labs/agentic-agent-node/tests/agent_payment_mocks/repositories_mocks"
	payment_interfaces_mocks "github.com/6022-labs/agentic-agent-node/tests/agent_payment_mocks/services_mocks/interfaces_mocks"
	"github.com/6022-labs/agentic-agent-node/tests/common_tests"
	"github.com/6022-labs/agentic-agent-node/tests/common_tests/model_builders"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
	"go.uber.org/zap"
)

type WhenEvaluatingX402AccessTestingSuite struct {
	sut *services.X402PaymentReceivePolicyEvaluator

	mockPaymentReceivePolicyProvider           *payment_interfaces_mocks.MockPaymentReceivePolicyProviderInterface
	mockPaymentReceiveRuleRepository           *payment_repositories_mocks.MockPaymentReceiveRuleRepositoryInterface
	mockPaymentReceiveRuleX402ConfigRepository *payment_repositories_mocks.MockPaymentReceiveRuleX402ConfigRepositoryInterface
	mockPaymentReceiveRulePriceRepository      *payment_repositories_mocks.MockPaymentReceiveRulePriceRepositoryInterface
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
	mockPaymentReceiveRulePriceRepository := payment_repositories_mocks.NewMockPaymentReceiveRulePriceRepositoryInterface(
		mockController,
	)

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
			zap.NewNop(),
			mockPaymentReceivePolicyProvider,
			mockPaymentReceiveRuleRepository,
			mockPaymentReceiveRuleX402ConfigRepository,
			mockPaymentReceiveRulePriceRepository,
			services.NewX402PaymentVerifier(),
			facilitators,
		),

		mockPaymentReceivePolicyProvider:           mockPaymentReceivePolicyProvider,
		mockPaymentReceiveRuleRepository:           mockPaymentReceiveRuleRepository,
		mockPaymentReceiveRuleX402ConfigRepository: mockPaymentReceiveRuleX402ConfigRepository,
		mockPaymentReceiveRulePriceRepository:      mockPaymentReceiveRulePriceRepository,
		callerKey:                                  callerKey,
	}
}

// x402TestRule is one token's rule with its config and prices.
type x402TestRule struct {
	config *models.PaymentReceiveRuleX402Config
	prices []models.PaymentReceiveRulePrice
}

// expectRules stubs the policy and one x402 rule per entry.
func (s *WhenEvaluatingX402AccessTestingSuite) expectRules(entries ...x402TestRule) {
	agent := model_builders.NewAgentBuilder().Build()
	policy := model_builders.NewPaymentReceivePolicyBuilder().WithAgentId(agent.Id).Build()

	s.mockPaymentReceivePolicyProvider.EXPECT().Ensure(gomock.Any()).Return(policy, agent, nil).Times(1)

	rules := make([]models.PaymentReceiveRule, 0, len(entries))
	for i := range entries {
		rule := model_builders.NewPaymentReceiveRuleBuilder().WithPolicyId(policy.Id).Build()
		rule.Method = enums.PaymentMethodX402
		rules = append(rules, *rule)

		config := entries[i].config
		config.PaymentReceiveRuleId = rule.Id
		s.mockPaymentReceiveRuleX402ConfigRepository.EXPECT().GetByRuleId(gomock.Any(), rule.Id).Return(config, nil)
		s.mockPaymentReceiveRulePriceRepository.EXPECT().
			ListByRuleId(gomock.Any(), rule.Id).
			Return(entries[i].prices, nil)
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

func publicRule(price string) x402TestRule {
	return scopedRule(price, ".*")
}

func scopedRule(price string, patterns ...string) x402TestRule {
	return x402TestRule{
		config: x402TestAcceptOption(),
		prices: []models.PaymentReceiveRulePrice{{Price: price, AddressPatterns: patterns}},
	}
}

func TestWhenEvaluatingX402Access(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

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

	t.Run("Given listing the rules fails", func(t *testing.T) {
		t.Parallel()

		t.Run("Should return an InternalError rather than a refusal", func(t *testing.T) {
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

	t.Run("Given a rule whose x402 config cannot be read", func(t *testing.T) {
		t.Parallel()

		t.Run("Should return an InternalError rather than a refusal", func(t *testing.T) {
			t.Parallel()

			suite := WhenEvaluatingX402AccessBeforeEach(t)
			agent := model_builders.NewAgentBuilder().Build()
			policy := model_builders.NewPaymentReceivePolicyBuilder().WithAgentId(agent.Id).Build()
			rule := model_builders.NewPaymentReceiveRuleBuilder().WithPolicyId(policy.Id).Build()
			rule.Method = enums.PaymentMethodX402
			suite.mockPaymentReceivePolicyProvider.EXPECT().Ensure(gomock.Any()).Return(policy, agent, nil)
			suite.mockPaymentReceiveRuleRepository.EXPECT().
				ListByPolicyId(gomock.Any(), policy.Id).
				Return([]models.PaymentReceiveRule{*rule}, nil)
			suite.mockPaymentReceiveRuleX402ConfigRepository.EXPECT().
				GetByRuleId(gomock.Any(), rule.Id).
				Return(nil, assert.AnError)

			decision, err := suite.sut.Evaluate(ctx, &payment_interfaces.PaymentReceiveAccessCredentials{})

			assert.Nil(t, decision)
			var internalError *apperrors.InternalError
			assert.ErrorAs(t, err, &internalError)
			assert.Equal(t, "access_evaluation_failed", internalError.Code)
		})
	})

	t.Run("Given listing a rule's prices fails", func(t *testing.T) {
		t.Parallel()

		t.Run("Should return an InternalError rather than a refusal", func(t *testing.T) {
			t.Parallel()

			suite := WhenEvaluatingX402AccessBeforeEach(t)
			agent := model_builders.NewAgentBuilder().Build()
			policy := model_builders.NewPaymentReceivePolicyBuilder().WithAgentId(agent.Id).Build()
			rule := model_builders.NewPaymentReceiveRuleBuilder().WithPolicyId(policy.Id).Build()
			rule.Method = enums.PaymentMethodX402
			suite.mockPaymentReceivePolicyProvider.EXPECT().Ensure(gomock.Any()).Return(policy, agent, nil)
			suite.mockPaymentReceiveRuleRepository.EXPECT().
				ListByPolicyId(gomock.Any(), policy.Id).
				Return([]models.PaymentReceiveRule{*rule}, nil)
			suite.mockPaymentReceiveRuleX402ConfigRepository.EXPECT().
				GetByRuleId(gomock.Any(), rule.Id).
				Return(x402TestAcceptOption(), nil)
			suite.mockPaymentReceiveRulePriceRepository.EXPECT().
				ListByRuleId(gomock.Any(), rule.Id).
				Return(nil, assert.AnError)

			decision, err := suite.sut.Evaluate(ctx, &payment_interfaces.PaymentReceiveAccessCredentials{})

			assert.Nil(t, decision)
			var internalError *apperrors.InternalError
			assert.ErrorAs(t, err, &internalError)
			assert.Equal(t, "access_evaluation_failed", internalError.Code)
		})
	})

	t.Run("Given a JWT-proven caller a free rule covers", func(t *testing.T) {
		t.Parallel()

		t.Run("Should grant free access without a signature", func(t *testing.T) {
			t.Parallel()

			suite := WhenEvaluatingX402AccessBeforeEach(t)
			caller := crypto.PubkeyToAddress(suite.callerKey.PublicKey)
			suite.expectRules(scopedRule("0", strings.ToLower(caller.Hex())))

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
			suite.expectRules(scopedRule("0", strings.ToLower(common_tests.RandomEthAddress().Hex())))

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
			scoped := scopedRule("10000", strings.ToLower(common_tests.RandomEthAddress().Hex()))
			// Distinct payTo so the advertised entry identifies which rule was chosen.
			scoped.config.PayTo = common_tests.RandomEthAddress().Hex()
			suite.expectRules(publicRule("10000"), scoped)

			decision, err := suite.sut.Evaluate(ctx, &payment_interfaces.PaymentReceiveAccessCredentials{})

			assert.NoError(t, err)
			assert.Equal(t, enums.PaymentReceiveAccessOutcomePaymentRequired, decision.Outcome)
			// The scoped rule stays hidden: advertising it would leak its address patterns.
			if assert.Len(t, decision.Accepts, 1) {
				assert.Equal(t, x402TestAcceptOption().PayTo, decision.Accepts[0].PayTo)
				assert.Equal(t, 90, decision.Accepts[0].MaxTimeoutSeconds)
				// eip3009 needs the token's domain to sign; permit2 rules carry none.
				assert.Equal(t, "eip3009", decision.Accepts[0].Extra["assetTransferMethod"])
			}
		})
	})

	t.Run("Given a rule with a public price and a private one", func(t *testing.T) {
		t.Parallel()

		initSuite := func(suite *WhenEvaluatingX402AccessTestingSuite) {
			rule := publicRule("10000")
			// Stored lowercased as the one-rule-per-token key.
			rule.config.Asset = strings.ToLower(rule.config.Asset)
			rule.prices = append(rule.prices, models.PaymentReceiveRulePrice{
				Price:           "0",
				AddressPatterns: []string{strings.ToLower(common_tests.RandomEthAddress().Hex())},
			})
			suite.expectRules(rule)
		}

		t.Run("Should advertise only the public price", func(t *testing.T) {
			t.Parallel()

			suite := WhenEvaluatingX402AccessBeforeEach(t)
			initSuite(suite)

			decision, err := suite.sut.Evaluate(ctx, &payment_interfaces.PaymentReceiveAccessCredentials{})

			assert.NoError(t, err)
			if assert.Len(t, decision.Accepts, 1) {
				assert.Equal(t, "10000", decision.Accepts[0].Amount)
			}
		})

		t.Run("Should advertise the token address checksummed", func(t *testing.T) {
			t.Parallel()

			suite := WhenEvaluatingX402AccessBeforeEach(t)
			initSuite(suite)

			decision, err := suite.sut.Evaluate(ctx, &payment_interfaces.PaymentReceiveAccessCredentials{})

			assert.NoError(t, err)
			if assert.Len(t, decision.Accepts, 1) {
				assert.Equal(t, x402TestAcceptOption().Asset, decision.Accepts[0].Asset)
			}
		})
	})

	t.Run("Given a paid rule on a network with no facilitator", func(t *testing.T) {
		t.Parallel()

		t.Run("Should not advertise it, since it could never settle", func(t *testing.T) {
			t.Parallel()

			suite := WhenEvaluatingX402AccessBeforeEach(t)
			unsettleable := publicRule("10000")
			unsettleable.config.Network = "eip155:1"
			suite.expectRules(unsettleable)

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
			suite.expectRules(publicRule("10000"))

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
			signedAgainst := x402TestAcceptOption()
			signedAgainst.PayTo = common_tests.RandomEthAddress().Hex()
			header := suite.signedHeader(t, signedAgainst)

			suite.expectRules(publicRule("10000"))

			decision, err := suite.sut.Evaluate(ctx, &payment_interfaces.PaymentReceiveAccessCredentials{
				PaymentSignature: header,
			})

			assert.NoError(t, err)
			assert.Equal(t, "no_matching_payment_requirements", decision.ErrorCode)
		})
	})

	t.Run("Given an authorization outside its validity window", func(t *testing.T) {
		t.Parallel()

		t.Run("Should refuse as expired", func(t *testing.T) {
			t.Parallel()

			suite := WhenEvaluatingX402AccessBeforeEach(t)
			rule := publicRule("10000")
			from := crypto.PubkeyToAddress(suite.callerKey.PublicKey)
			authorization := x402TestAuthorization(from.Hex(), rule.config)
			authorization.ValidBefore = fmt.Sprintf("%d", time.Now().Unix()-30)
			header, encodeErr := x402.EncodeBase64Json(signX402Payment(t, suite.callerKey, authorization, rule.config))
			assert.NoError(t, encodeErr)
			suite.expectRules(rule)

			decision, err := suite.sut.Evaluate(ctx, &payment_interfaces.PaymentReceiveAccessCredentials{
				PaymentSignature: header,
			})

			assert.NoError(t, err)
			assert.Equal(t, "authorization_expired", decision.ErrorCode)
		})
	})

	t.Run("Given a signature that does not recover to the authorization's payer", func(t *testing.T) {
		t.Parallel()

		t.Run("Should refuse as an invalid signature", func(t *testing.T) {
			t.Parallel()

			suite := WhenEvaluatingX402AccessBeforeEach(t)
			rule := publicRule("10000")
			payload := signX402Payment(t, suite.callerKey, x402TestAuthorization(
				common_tests.RandomEthAddress().Hex(), rule.config,
			), rule.config)
			header, encodeErr := x402.EncodeBase64Json(payload)
			assert.NoError(t, encodeErr)
			suite.expectRules(rule)

			decision, err := suite.sut.Evaluate(ctx, &payment_interfaces.PaymentReceiveAccessCredentials{
				PaymentSignature: header,
			})

			assert.NoError(t, err)
			assert.Equal(t, "invalid_payment_signature", decision.ErrorCode)
		})
	})

	t.Run("Given a payload carrying a second, unsigned authorization", func(t *testing.T) {
		t.Parallel()

		// A payload signs exactly one authorization. Attaching an unsigned permit2 blob beside a
		// genuine eip3009 one used to hijack the amount check, since the method-neutral reader
		// takes permit2 first while the eip3009 verifier never looks at it.
		initSuite := func(suite *WhenEvaluatingX402AccessTestingSuite) string {
			rule := publicRule(testPrice)
			config := rule.config

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

			suite.expectRules(rule)
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

	t.Run("Given a signer no rule entitles", func(t *testing.T) {
		t.Parallel()

		t.Run("Should refuse as unmatched", func(t *testing.T) {
			t.Parallel()

			suite := WhenEvaluatingX402AccessBeforeEach(t)
			scoped := scopedRule("10000", strings.ToLower(common_tests.RandomEthAddress().Hex()))
			header := suite.signedHeader(t, scoped.config)
			suite.expectRules(scoped)

			decision, err := suite.sut.Evaluate(ctx, &payment_interfaces.PaymentReceiveAccessCredentials{
				PaymentSignature: header,
			})

			assert.NoError(t, err)
			assert.Equal(t, "no_matching_payment_requirements", decision.ErrorCode)
		})
	})

	t.Run("Given the signed amount is below the rule's price", func(t *testing.T) {
		t.Parallel()

		t.Run("Should refuse as insufficient", func(t *testing.T) {
			t.Parallel()

			suite := WhenEvaluatingX402AccessBeforeEach(t)
			rule := publicRule("999999999")
			// The signature is worth testPrice; the rule now asks for more.
			header := suite.signedHeader(t, rule.config)
			suite.expectRules(rule)

			decision, err := suite.sut.Evaluate(ctx, &payment_interfaces.PaymentReceiveAccessCredentials{
				PaymentSignature: header,
			})

			assert.NoError(t, err)
			assert.Equal(t, "insufficient_amount", decision.ErrorCode)
		})
	})

	t.Run("Given another token's rule has a free price for the payer", func(t *testing.T) {
		t.Parallel()

		// Only the token the payer signed for can entitle them; another token's free price must not leak in.
		t.Run("Should charge the signed token's price instead of granting free access", func(t *testing.T) {
			t.Parallel()

			suite := WhenEvaluatingX402AccessBeforeEach(t)
			caller := crypto.PubkeyToAddress(suite.callerKey.PublicKey)
			signed := publicRule(testPrice)
			// Same network and payee, so both rules are candidates until the signature is checked.
			other := scopedRule("0", strings.ToLower(caller.Hex()))
			other.config.Asset = common_tests.RandomEthAddress().Hex()
			header := suite.signedHeader(t, signed.config)
			suite.expectRules(signed, other)

			decision, err := suite.sut.Evaluate(ctx, &payment_interfaces.PaymentReceiveAccessCredentials{
				PaymentSignature: header,
			})

			assert.NoError(t, err)
			assert.Equal(t, enums.PaymentReceiveAccessOutcomePayable, decision.Outcome)
			if assert.NotNil(t, decision.Settlement) {
				assert.Equal(t, testPrice, decision.Settlement.Requirements.Amount)
				assert.Equal(t, x402TestAcceptOption().Asset, decision.Settlement.Requirements.Asset)
			}
		})
	})

	t.Run("Given the payer is entitled to a free price as well", func(t *testing.T) {
		t.Parallel()

		// The cheapest entitled price wins, so a payer covered by a free price is never charged.
		t.Run("Should grant free access rather than settle", func(t *testing.T) {
			t.Parallel()

			suite := WhenEvaluatingX402AccessBeforeEach(t)
			caller := crypto.PubkeyToAddress(suite.callerKey.PublicKey)
			rule := publicRule(testPrice)
			rule.prices = append(rule.prices, models.PaymentReceiveRulePrice{
				Price:           "0",
				AddressPatterns: []string{strings.ToLower(caller.Hex())},
			})
			header := suite.signedHeader(t, rule.config)
			suite.expectRules(rule)

			decision, err := suite.sut.Evaluate(ctx, &payment_interfaces.PaymentReceiveAccessCredentials{
				PaymentSignature: header,
			})

			assert.NoError(t, err)
			assert.Equal(t, enums.PaymentReceiveAccessOutcomeFree, decision.Outcome)
			assert.Nil(t, decision.Settlement)
		})
	})

	t.Run("Given a signature that covers a public paid rule", func(t *testing.T) {
		t.Parallel()

		t.Run("Should be payable and carry what settles it", func(t *testing.T) {
			t.Parallel()

			suite := WhenEvaluatingX402AccessBeforeEach(t)
			rule := publicRule(testPrice)
			header := suite.signedHeader(t, rule.config)
			suite.expectRules(rule)

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
}
