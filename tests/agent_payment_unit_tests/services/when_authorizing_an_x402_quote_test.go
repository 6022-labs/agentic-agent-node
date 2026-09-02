package services_test

import (
	"context"
	"math/big"
	"testing"

	"github.com/6022-labs/agentic-agent-node/src/agent_payment/models"
	"github.com/6022-labs/agentic-agent-node/src/agent_payment/models/enums"
	"github.com/6022-labs/agentic-agent-node/src/agent_payment/services"
	payment_interfaces "github.com/6022-labs/agentic-agent-node/src/agent_payment/services/interfaces"
	apperrors "github.com/6022-labs/agentic-agent-node/src/common/errors"
	payment_repositories_mocks "github.com/6022-labs/agentic-agent-node/tests/agent_payment_mocks/repositories_mocks"
	payment_interfaces_mocks "github.com/6022-labs/agentic-agent-node/tests/agent_payment_mocks/services_mocks/interfaces_mocks"
	"github.com/6022-labs/agentic-agent-node/tests/common_tests/model_builders"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

type WhenAuthorizingAnX402QuoteTestingSuite struct {
	sut *services.X402PaymentSpendPolicyEvaluator

	mockPaymentSpendPolicyProvider           *payment_interfaces_mocks.MockPaymentSpendPolicyProviderInterface
	mockPaymentSpendRuleRepository           *payment_repositories_mocks.MockPaymentSpendRuleRepositoryInterface
	mockPaymentSpendRuleX402ConfigRepository *payment_repositories_mocks.MockPaymentSpendRuleX402ConfigRepositoryInterface
}

func WhenAuthorizingAnX402QuoteBeforeEach(t *testing.T) *WhenAuthorizingAnX402QuoteTestingSuite {
	mockController := gomock.NewController(t)

	mockPaymentSpendPolicyProvider := payment_interfaces_mocks.NewMockPaymentSpendPolicyProviderInterface(
		mockController,
	)
	mockPaymentSpendRuleRepository := payment_repositories_mocks.NewMockPaymentSpendRuleRepositoryInterface(
		mockController,
	)
	mockPaymentSpendRuleX402ConfigRepository :=
		payment_repositories_mocks.NewMockPaymentSpendRuleX402ConfigRepositoryInterface(mockController)

	return &WhenAuthorizingAnX402QuoteTestingSuite{
		sut: services.NewX402PaymentSpendPolicyEvaluator(
			mockPaymentSpendPolicyProvider,
			mockPaymentSpendRuleRepository,
			mockPaymentSpendRuleX402ConfigRepository,
		),

		mockPaymentSpendPolicyProvider:           mockPaymentSpendPolicyProvider,
		mockPaymentSpendRuleRepository:           mockPaymentSpendRuleRepository,
		mockPaymentSpendRuleX402ConfigRepository: mockPaymentSpendRuleX402ConfigRepository,
	}
}

func authorizedQuote() *payment_interfaces.X402PaymentQuote {
	return &payment_interfaces.X402PaymentQuote{
		Network:             "eip155:80002",
		Asset:               "0x41E94Eb019C0762f9Bfcf9Fb1E58725BfB0e7582",
		PayTo:               "0xf39Fd6e51aad88F6F4ce6aB8827279cffFb92266",
		Amount:              "10000",
		AssetTransferMethod: enums.AssetTransferMethodEip3009,
	}
}

func TestWhenAuthorizingAnX402Quote(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	t.Run("Given a catch-all rule whose cap covers the amount", func(t *testing.T) {
		t.Parallel()

		initSuite := func(suite *WhenAuthorizingAnX402QuoteTestingSuite) {
			policy := model_builders.NewPaymentSpendPolicyBuilder().Build()
			rule := model_builders.NewPaymentSpendRuleBuilder().WithPolicyId(policy.Id).Build()
			rule.MaxAmount = "10000"

			suite.mockPaymentSpendPolicyProvider.EXPECT().Ensure(ctx).Return(policy, nil).Times(1)
			suite.mockPaymentSpendRuleRepository.EXPECT().
				ListByPolicyId(ctx, policy.Id).
				Return([]models.PaymentSpendRule{*rule}, nil).
				Times(1)
			suite.mockPaymentSpendRuleX402ConfigRepository.EXPECT().
				GetByRuleId(ctx, rule.Id).
				Return(model_builders.NewPaymentSpendRuleX402ConfigBuilder().WithPaymentSpendRuleId(rule.Id).Build(), nil).
				Times(1)
		}

		// The cap is inclusive: an amount equal to it is still within the rule.
		t.Run("Should authorize an amount equal to the cap", func(t *testing.T) {
			t.Parallel()

			suite := WhenAuthorizingAnX402QuoteBeforeEach(t)
			initSuite(suite)

			authorized, err := suite.sut.IsAuthorized(ctx, authorizedQuote(), big.NewInt(10000))

			assert.NoError(t, err)
			assert.True(t, authorized)
		})

		t.Run("Should refuse an amount above the cap", func(t *testing.T) {
			t.Parallel()

			suite := WhenAuthorizingAnX402QuoteBeforeEach(t)
			initSuite(suite)

			authorized, err := suite.sut.IsAuthorized(ctx, authorizedQuote(), big.NewInt(10001))

			assert.NoError(t, err)
			assert.False(t, authorized)
		})
	})

	t.Run("Given two rules match the quote", func(t *testing.T) {
		t.Parallel()

		// The effective cap is the minimum, so a narrow rule cannot be widened by adding a broad one.
		initSuite := func(suite *WhenAuthorizingAnX402QuoteTestingSuite) {
			policy := model_builders.NewPaymentSpendPolicyBuilder().Build()
			generous := model_builders.NewPaymentSpendRuleBuilder().WithPolicyId(policy.Id).Build()
			generous.MaxAmount = "1000000"
			strict := model_builders.NewPaymentSpendRuleBuilder().WithPolicyId(policy.Id).Build()
			strict.MaxAmount = "500"

			suite.mockPaymentSpendPolicyProvider.EXPECT().Ensure(ctx).Return(policy, nil).Times(1)
			suite.mockPaymentSpendRuleRepository.EXPECT().
				ListByPolicyId(ctx, policy.Id).
				Return([]models.PaymentSpendRule{*generous, *strict}, nil).
				Times(1)
			for _, rule := range []*models.PaymentSpendRule{generous, strict} {
				suite.mockPaymentSpendRuleX402ConfigRepository.EXPECT().
					GetByRuleId(ctx, rule.Id).
					Return(
						model_builders.NewPaymentSpendRuleX402ConfigBuilder().WithPaymentSpendRuleId(rule.Id).Build(),
						nil,
					).
					Times(1)
			}
		}

		t.Run("Should refuse an amount the strictest rule does not cover", func(t *testing.T) {
			t.Parallel()

			suite := WhenAuthorizingAnX402QuoteBeforeEach(t)
			initSuite(suite)

			authorized, err := suite.sut.IsAuthorized(ctx, authorizedQuote(), big.NewInt(501))

			assert.NoError(t, err)
			assert.False(t, authorized)
		})

		t.Run("Should authorize an amount the strictest rule covers", func(t *testing.T) {
			t.Parallel()

			suite := WhenAuthorizingAnX402QuoteBeforeEach(t)
			initSuite(suite)

			authorized, err := suite.sut.IsAuthorized(ctx, authorizedQuote(), big.NewInt(500))

			assert.NoError(t, err)
			assert.True(t, authorized)
		})
	})

	t.Run("Given no rule matches the quote", func(t *testing.T) {
		t.Parallel()

		initSuite := func(suite *WhenAuthorizingAnX402QuoteTestingSuite) {
			policy := model_builders.NewPaymentSpendPolicyBuilder().Build()
			rule := model_builders.NewPaymentSpendRuleBuilder().WithPolicyId(policy.Id).Build()

			suite.mockPaymentSpendPolicyProvider.EXPECT().Ensure(ctx).Return(policy, nil).Times(1)
			suite.mockPaymentSpendRuleRepository.EXPECT().
				ListByPolicyId(ctx, policy.Id).
				Return([]models.PaymentSpendRule{*rule}, nil).
				Times(1)
			suite.mockPaymentSpendRuleX402ConfigRepository.EXPECT().
				GetByRuleId(ctx, rule.Id).
				Return(
					model_builders.NewPaymentSpendRuleX402ConfigBuilder().
						WithPaymentSpendRuleId(rule.Id).
						WithPayToPatterns([]string{"0x0000000000000000000000000000000000000000"}).
						Build(),
					nil,
				).
				Times(1)
		}

		// Deny by default: an unmatched quote is one the owner never described.
		t.Run("Should refuse rather than fall through", func(t *testing.T) {
			t.Parallel()

			suite := WhenAuthorizingAnX402QuoteBeforeEach(t)
			initSuite(suite)

			authorized, err := suite.sut.IsAuthorized(ctx, authorizedQuote(), big.NewInt(1))

			assert.NoError(t, err)
			assert.False(t, authorized)
		})
	})

	t.Run("Given the policy carries no rules at all", func(t *testing.T) {
		t.Parallel()

		initSuite := func(suite *WhenAuthorizingAnX402QuoteTestingSuite) {
			policy := model_builders.NewPaymentSpendPolicyBuilder().Build()

			suite.mockPaymentSpendPolicyProvider.EXPECT().Ensure(ctx).Return(policy, nil).Times(1)
			suite.mockPaymentSpendRuleRepository.EXPECT().ListByPolicyId(ctx, policy.Id).Return(nil, nil).Times(1)
		}

		t.Run("Should refuse", func(t *testing.T) {
			t.Parallel()

			suite := WhenAuthorizingAnX402QuoteBeforeEach(t)
			initSuite(suite)

			authorized, err := suite.sut.IsAuthorized(ctx, authorizedQuote(), big.NewInt(1))

			assert.NoError(t, err)
			assert.False(t, authorized)
		})
	})

	t.Run("Given the matching rule caps at zero", func(t *testing.T) {
		t.Parallel()

		// "0" is the deny idiom: it still matches, so a free quote passes and any priced one does not.
		initSuite := func(suite *WhenAuthorizingAnX402QuoteTestingSuite) {
			policy := model_builders.NewPaymentSpendPolicyBuilder().Build()
			rule := model_builders.NewPaymentSpendRuleBuilder().WithPolicyId(policy.Id).Build()
			rule.MaxAmount = "0"

			suite.mockPaymentSpendPolicyProvider.EXPECT().Ensure(ctx).Return(policy, nil).Times(1)
			suite.mockPaymentSpendRuleRepository.EXPECT().
				ListByPolicyId(ctx, policy.Id).
				Return([]models.PaymentSpendRule{*rule}, nil).
				Times(1)
			suite.mockPaymentSpendRuleX402ConfigRepository.EXPECT().
				GetByRuleId(ctx, rule.Id).
				Return(model_builders.NewPaymentSpendRuleX402ConfigBuilder().WithPaymentSpendRuleId(rule.Id).Build(), nil).
				Times(1)
		}

		t.Run("Should refuse any priced quote", func(t *testing.T) {
			t.Parallel()

			suite := WhenAuthorizingAnX402QuoteBeforeEach(t)
			initSuite(suite)

			authorized, err := suite.sut.IsAuthorized(ctx, authorizedQuote(), big.NewInt(1))

			assert.NoError(t, err)
			assert.False(t, authorized)
		})

		t.Run("Should still authorize a free quote", func(t *testing.T) {
			t.Parallel()

			suite := WhenAuthorizingAnX402QuoteBeforeEach(t)
			initSuite(suite)

			authorized, err := suite.sut.IsAuthorized(ctx, authorizedQuote(), big.NewInt(0))

			assert.NoError(t, err)
			assert.True(t, authorized)
		})
	})

	t.Run("Given the only matching rule belongs to another method", func(t *testing.T) {
		t.Parallel()

		initSuite := func(suite *WhenAuthorizingAnX402QuoteTestingSuite) {
			policy := model_builders.NewPaymentSpendPolicyBuilder().Build()
			rule := model_builders.NewPaymentSpendRuleBuilder().WithPolicyId(policy.Id).Build()
			rule.Method = "lightning"

			suite.mockPaymentSpendPolicyProvider.EXPECT().Ensure(ctx).Return(policy, nil).Times(1)
			suite.mockPaymentSpendRuleRepository.EXPECT().
				ListByPolicyId(ctx, policy.Id).
				Return([]models.PaymentSpendRule{*rule}, nil).
				Times(1)
		}

		t.Run("Should refuse without reading its x402 matcher", func(t *testing.T) {
			t.Parallel()

			suite := WhenAuthorizingAnX402QuoteBeforeEach(t)
			initSuite(suite)

			authorized, err := suite.sut.IsAuthorized(ctx, authorizedQuote(), big.NewInt(1))

			assert.NoError(t, err)
			assert.False(t, authorized)
		})
	})

	t.Run("Given a rule's matcher row is missing", func(t *testing.T) {
		t.Parallel()

		initSuite := func(suite *WhenAuthorizingAnX402QuoteTestingSuite) {
			policy := model_builders.NewPaymentSpendPolicyBuilder().Build()
			rule := model_builders.NewPaymentSpendRuleBuilder().WithPolicyId(policy.Id).Build()

			suite.mockPaymentSpendPolicyProvider.EXPECT().Ensure(ctx).Return(policy, nil).Times(1)
			suite.mockPaymentSpendRuleRepository.EXPECT().
				ListByPolicyId(ctx, policy.Id).
				Return([]models.PaymentSpendRule{*rule}, nil).
				Times(1)
			suite.mockPaymentSpendRuleX402ConfigRepository.EXPECT().GetByRuleId(ctx, rule.Id).Return(nil, nil).Times(1)
		}

		// A rule that describes nothing grants nothing, rather than everything.
		t.Run("Should treat it as granting nothing", func(t *testing.T) {
			t.Parallel()

			suite := WhenAuthorizingAnX402QuoteBeforeEach(t)
			initSuite(suite)

			authorized, err := suite.sut.IsAuthorized(ctx, authorizedQuote(), big.NewInt(1))

			assert.NoError(t, err)
			assert.False(t, authorized)
		})
	})

	t.Run("Given a matching rule carries an unusable cap", func(t *testing.T) {
		t.Parallel()

		initSuite := func(suite *WhenAuthorizingAnX402QuoteTestingSuite) {
			policy := model_builders.NewPaymentSpendPolicyBuilder().Build()
			rule := model_builders.NewPaymentSpendRuleBuilder().WithPolicyId(policy.Id).Build()
			rule.MaxAmount = "not-a-number"

			suite.mockPaymentSpendPolicyProvider.EXPECT().Ensure(ctx).Return(policy, nil).Times(1)
			suite.mockPaymentSpendRuleRepository.EXPECT().
				ListByPolicyId(ctx, policy.Id).
				Return([]models.PaymentSpendRule{*rule}, nil).
				Times(1)
			suite.mockPaymentSpendRuleX402ConfigRepository.EXPECT().
				GetByRuleId(ctx, rule.Id).
				Return(model_builders.NewPaymentSpendRuleX402ConfigBuilder().WithPaymentSpendRuleId(rule.Id).Build(), nil).
				Times(1)
		}

		// Refusing loudly beats silently ignoring the rule that was meant to bound the spend.
		t.Run("Should fail rather than skip the rule", func(t *testing.T) {
			t.Parallel()

			suite := WhenAuthorizingAnX402QuoteBeforeEach(t)
			initSuite(suite)

			authorized, err := suite.sut.IsAuthorized(ctx, authorizedQuote(), big.NewInt(1))

			assert.False(t, authorized)
			var internalError *apperrors.InternalError
			assert.ErrorAs(t, err, &internalError)
		})
	})

	t.Run("Given materializing the policy fails", func(t *testing.T) {
		t.Parallel()

		initSuite := func(suite *WhenAuthorizingAnX402QuoteTestingSuite) {
			suite.mockPaymentSpendPolicyProvider.EXPECT().
				Ensure(ctx).
				Return(nil, apperrors.NewInternalError("ensure_failed", "")).
				Times(1)
		}

		t.Run("Should propagate the provider's error", func(t *testing.T) {
			t.Parallel()

			suite := WhenAuthorizingAnX402QuoteBeforeEach(t)
			initSuite(suite)

			authorized, err := suite.sut.IsAuthorized(ctx, authorizedQuote(), big.NewInt(1))

			assert.False(t, authorized)
			var internalError *apperrors.InternalError
			assert.ErrorAs(t, err, &internalError)
		})
	})

	t.Run("Given listing the rules fails", func(t *testing.T) {
		t.Parallel()

		initSuite := func(suite *WhenAuthorizingAnX402QuoteTestingSuite) {
			policy := model_builders.NewPaymentSpendPolicyBuilder().Build()

			suite.mockPaymentSpendPolicyProvider.EXPECT().Ensure(ctx).Return(policy, nil).Times(1)
			suite.mockPaymentSpendRuleRepository.EXPECT().
				ListByPolicyId(ctx, policy.Id).
				Return(nil, assert.AnError).
				Times(1)
		}

		t.Run("Should return an internal error", func(t *testing.T) {
			t.Parallel()

			suite := WhenAuthorizingAnX402QuoteBeforeEach(t)
			initSuite(suite)

			authorized, err := suite.sut.IsAuthorized(ctx, authorizedQuote(), big.NewInt(1))

			assert.False(t, authorized)
			var internalError *apperrors.InternalError
			assert.ErrorAs(t, err, &internalError)
		})
	})

	t.Run("Given reading a matcher fails", func(t *testing.T) {
		t.Parallel()

		initSuite := func(suite *WhenAuthorizingAnX402QuoteTestingSuite) {
			policy := model_builders.NewPaymentSpendPolicyBuilder().Build()
			rule := model_builders.NewPaymentSpendRuleBuilder().WithPolicyId(policy.Id).Build()

			suite.mockPaymentSpendPolicyProvider.EXPECT().Ensure(ctx).Return(policy, nil).Times(1)
			suite.mockPaymentSpendRuleRepository.EXPECT().
				ListByPolicyId(ctx, policy.Id).
				Return([]models.PaymentSpendRule{*rule}, nil).
				Times(1)
			suite.mockPaymentSpendRuleX402ConfigRepository.EXPECT().
				GetByRuleId(ctx, rule.Id).
				Return(nil, assert.AnError).
				Times(1)
		}

		t.Run("Should return an internal error rather than authorize", func(t *testing.T) {
			t.Parallel()

			suite := WhenAuthorizingAnX402QuoteBeforeEach(t)
			initSuite(suite)

			authorized, err := suite.sut.IsAuthorized(ctx, authorizedQuote(), big.NewInt(1))

			assert.False(t, authorized)
			var internalError *apperrors.InternalError
			assert.ErrorAs(t, err, &internalError)
		})
	})
}
