package services_test

import (
	"context"
	"math/big"
	"strings"
	"testing"
	"time"

	"github.com/6022-labs/agentic-agent-node/src/agent_payment/models"
	"github.com/6022-labs/agentic-agent-node/src/agent_payment/models/enums"
	"github.com/6022-labs/agentic-agent-node/src/agent_payment/services"
	payment_interfaces "github.com/6022-labs/agentic-agent-node/src/agent_payment/services/interfaces"
	apperrors "github.com/6022-labs/agentic-agent-node/src/common/errors"
	runtime_common_repositories "github.com/6022-labs/agentic-agent-node/src/runtime_common/repositories"
	payment_repositories_mocks "github.com/6022-labs/agentic-agent-node/tests/agent_payment_mocks/repositories_mocks"
	payment_interfaces_mocks "github.com/6022-labs/agentic-agent-node/tests/agent_payment_mocks/services_mocks/interfaces_mocks"
	"github.com/6022-labs/agentic-agent-node/tests/common_tests"
	"github.com/6022-labs/agentic-agent-node/tests/common_tests/model_builders"
	runtime_common_mocks "github.com/6022-labs/agentic-agent-node/tests/runtime_common_mocks/services_mocks/interfaces_mocks"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

const quoteNetwork = "eip155:80002"

var (
	quoteAssetAddress = common_tests.RandomEthAddress()
	// The quote spells the asset checksummed; the rule stores it lowercased.
	quoteAsset       = quoteAssetAddress.Hex()
	storedQuoteAsset = strings.ToLower(quoteAssetAddress.Hex())
	quotePayTo       = common_tests.RandomEthAddress().Hex()
)

type authorizeUnitOfWorkContextKey struct{}

type WhenAuthorizingAnX402QuoteTestingSuite struct {
	sut *services.X402PaymentSpendPolicyEvaluator

	mockPaymentSpendPolicyProvider           *payment_interfaces_mocks.MockPaymentSpendPolicyProviderInterface
	mockPaymentSpendRuleRepository           *payment_repositories_mocks.MockPaymentSpendRuleRepositoryInterface
	mockPaymentSpendRuleX402ConfigRepository *payment_repositories_mocks.MockPaymentSpendRuleX402ConfigRepositoryInterface
	mockPaymentSpendRuleWindowRepository     *payment_repositories_mocks.MockPaymentSpendRuleWindowRepositoryInterface
	mockUnitOfWorkFactory                    *runtime_common_mocks.MockUnitOfWorkFactoryInterface
	mockUnitOfWork                           *runtime_common_mocks.MockUnitOfWorkInterface
	unitOfWorkCtx                            context.Context
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
	mockPaymentSpendRuleWindowRepository :=
		payment_repositories_mocks.NewMockPaymentSpendRuleWindowRepositoryInterface(mockController)
	mockUnitOfWorkFactory := runtime_common_mocks.NewMockUnitOfWorkFactoryInterface(mockController)
	// The real factory opens a transaction; here the scope carries a context distinct from the caller's.
	unitOfWorkCtx := context.WithValue(context.Background(), authorizeUnitOfWorkContextKey{}, "unit_of_work")
	mockUnitOfWork := runtime_common_mocks.NewMockUnitOfWorkInterface(mockController)
	mockUnitOfWork.EXPECT().Context().Return(unitOfWorkCtx).AnyTimes()
	mockUnitOfWork.EXPECT().Rollback().Return(nil).AnyTimes()

	return &WhenAuthorizingAnX402QuoteTestingSuite{
		sut: services.NewX402PaymentSpendPolicyEvaluator(
			mockPaymentSpendPolicyProvider,
			mockPaymentSpendRuleRepository,
			mockPaymentSpendRuleX402ConfigRepository,
			mockPaymentSpendRuleWindowRepository,
			mockUnitOfWorkFactory,
		),

		mockPaymentSpendPolicyProvider:           mockPaymentSpendPolicyProvider,
		mockPaymentSpendRuleRepository:           mockPaymentSpendRuleRepository,
		mockPaymentSpendRuleX402ConfigRepository: mockPaymentSpendRuleX402ConfigRepository,
		mockPaymentSpendRuleWindowRepository:     mockPaymentSpendRuleWindowRepository,
		mockUnitOfWorkFactory:                    mockUnitOfWorkFactory,
		mockUnitOfWork:                           mockUnitOfWork,
		unitOfWorkCtx:                            unitOfWorkCtx,
	}
}

func authorizedQuote() *payment_interfaces.X402PaymentQuote {
	return &payment_interfaces.X402PaymentQuote{
		Network:             quoteNetwork,
		Asset:               quoteAsset,
		PayTo:               quotePayTo,
		Amount:              "10000",
		AssetTransferMethod: enums.AssetTransferMethodEip3009,
	}
}

func TestWhenAuthorizingAnX402Quote(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	t.Run("Given ensuring the policy fails", func(t *testing.T) {
		t.Parallel()

		initSuite := func(suite *WhenAuthorizingAnX402QuoteTestingSuite) {
			suite.mockPaymentSpendPolicyProvider.EXPECT().
				Ensure(ctx).
				Return(nil, apperrors.NewInternalError("spend_policy_failed", "")).
				Times(1)
		}

		t.Run("Should return the error", func(t *testing.T) {
			t.Parallel()

			suite := WhenAuthorizingAnX402QuoteBeforeEach(t)
			initSuite(suite)

			_, err := suite.sut.Authorize(ctx, authorizedQuote(), big.NewInt(10000))

			var internalError *apperrors.InternalError
			assert.ErrorAs(t, err, &internalError)
			assert.Equal(t, "spend_policy_failed", internalError.Code)
		})
	})

	t.Run("Given the agent is not minted", func(t *testing.T) {
		t.Parallel()

		initSuite := func(suite *WhenAuthorizingAnX402QuoteTestingSuite) {
			suite.mockPaymentSpendPolicyProvider.EXPECT().
				Ensure(ctx).
				Return(nil, apperrors.NewConflictError("agent_not_minted", "")).
				Times(1)
		}

		t.Run("Should refuse the quote", func(t *testing.T) {
			t.Parallel()

			suite := WhenAuthorizingAnX402QuoteBeforeEach(t)
			initSuite(suite)

			reservation, err := suite.sut.Authorize(ctx, authorizedQuote(), big.NewInt(10000))

			assert.Nil(t, reservation)
			var unauthorizedError *apperrors.UnauthorizedError
			assert.ErrorAs(t, err, &unauthorizedError)
			assert.Equal(t, "spend_agent_not_minted", unauthorizedError.Code)
		})
	})

	t.Run("Given the payee is not on the policy's list", func(t *testing.T) {
		t.Parallel()

		policy := model_builders.NewPaymentSpendPolicyBuilder().WithPayToPatterns([]string{}).Build()

		initSuite := func(suite *WhenAuthorizingAnX402QuoteTestingSuite) {
			suite.mockPaymentSpendPolicyProvider.EXPECT().Ensure(ctx).Return(policy, nil).Times(1)
		}

		t.Run("Should refuse the quote", func(t *testing.T) {
			t.Parallel()

			suite := WhenAuthorizingAnX402QuoteBeforeEach(t)
			initSuite(suite)

			reservation, err := suite.sut.Authorize(ctx, authorizedQuote(), big.NewInt(10000))

			assert.Nil(t, reservation)
			var unauthorizedError *apperrors.UnauthorizedError
			assert.ErrorAs(t, err, &unauthorizedError)
			assert.Equal(t, "spend_payee_not_allowed", unauthorizedError.Code)
		})
	})

	t.Run("Given starting the unit of work fails", func(t *testing.T) {
		t.Parallel()

		policy := model_builders.NewPaymentSpendPolicyBuilder().Build()

		initSuite := func(suite *WhenAuthorizingAnX402QuoteTestingSuite) {
			suite.mockPaymentSpendPolicyProvider.EXPECT().Ensure(ctx).Return(policy, nil).Times(1)
			suite.mockUnitOfWorkFactory.EXPECT().Start(ctx).Return(nil, assert.AnError).Times(1)
		}

		t.Run("Should return an InternalError", func(t *testing.T) {
			t.Parallel()

			suite := WhenAuthorizingAnX402QuoteBeforeEach(t)
			initSuite(suite)

			_, err := suite.sut.Authorize(ctx, authorizedQuote(), big.NewInt(10000))

			var internalError *apperrors.InternalError
			assert.ErrorAs(t, err, &internalError)
			assert.Equal(t, "spend_policy_failed", internalError.Code)
		})
	})

	t.Run("Given reading the token's config fails", func(t *testing.T) {
		t.Parallel()

		policy := model_builders.NewPaymentSpendPolicyBuilder().Build()

		initSuite := func(suite *WhenAuthorizingAnX402QuoteTestingSuite) {
			suite.mockPaymentSpendPolicyProvider.EXPECT().Ensure(ctx).Return(policy, nil).Times(1)
			suite.mockUnitOfWorkFactory.EXPECT().Start(ctx).Return(suite.mockUnitOfWork, nil).Times(1)
			suite.mockPaymentSpendRuleX402ConfigRepository.EXPECT().
				GetByPolicyIdAndAsset(suite.unitOfWorkCtx, policy.Id, quoteNetwork, storedQuoteAsset).
				Return(nil, assert.AnError).
				Times(1)
		}

		t.Run("Should return an InternalError", func(t *testing.T) {
			t.Parallel()

			suite := WhenAuthorizingAnX402QuoteBeforeEach(t)
			initSuite(suite)

			_, err := suite.sut.Authorize(ctx, authorizedQuote(), big.NewInt(10000))

			var internalError *apperrors.InternalError
			assert.ErrorAs(t, err, &internalError)
			assert.Equal(t, "spend_policy_failed", internalError.Code)
		})
	})

	t.Run("Given no rule spends the quote's token", func(t *testing.T) {
		t.Parallel()

		policy := model_builders.NewPaymentSpendPolicyBuilder().Build()

		initSuite := func(suite *WhenAuthorizingAnX402QuoteTestingSuite) {
			suite.mockPaymentSpendPolicyProvider.EXPECT().Ensure(ctx).Return(policy, nil).Times(1)
			suite.mockUnitOfWorkFactory.EXPECT().Start(ctx).Return(suite.mockUnitOfWork, nil).Times(1)
			suite.mockPaymentSpendRuleX402ConfigRepository.EXPECT().
				GetByPolicyIdAndAsset(suite.unitOfWorkCtx, policy.Id, quoteNetwork, storedQuoteAsset).
				Return(nil, nil).
				Times(1)
		}

		// No match is a refusal, not a fallthrough: an unlisted token is one the owner never described.
		t.Run("Should refuse the quote", func(t *testing.T) {
			t.Parallel()

			suite := WhenAuthorizingAnX402QuoteBeforeEach(t)
			initSuite(suite)

			reservation, err := suite.sut.Authorize(ctx, authorizedQuote(), big.NewInt(10000))

			assert.Nil(t, reservation)
			var unauthorizedError *apperrors.UnauthorizedError
			assert.ErrorAs(t, err, &unauthorizedError)
			assert.Equal(t, "spend_token_has_no_rule", unauthorizedError.Code)
		})
	})

	t.Run("Given reading the token's rule fails", func(t *testing.T) {
		t.Parallel()

		policy := model_builders.NewPaymentSpendPolicyBuilder().Build()
		config := model_builders.NewPaymentSpendRuleX402ConfigBuilder().WithPolicyId(policy.Id).Build()

		initSuite := func(suite *WhenAuthorizingAnX402QuoteTestingSuite) {
			suite.mockPaymentSpendPolicyProvider.EXPECT().Ensure(ctx).Return(policy, nil).Times(1)
			suite.mockUnitOfWorkFactory.EXPECT().Start(ctx).Return(suite.mockUnitOfWork, nil).Times(1)
			suite.mockPaymentSpendRuleX402ConfigRepository.EXPECT().
				GetByPolicyIdAndAsset(suite.unitOfWorkCtx, policy.Id, quoteNetwork, storedQuoteAsset).
				Return(config, nil).
				Times(1)
			suite.mockPaymentSpendRuleRepository.EXPECT().
				GetById(suite.unitOfWorkCtx, config.PaymentSpendRuleId, runtime_common_repositories.ToLock).
				Return(nil, assert.AnError).
				Times(1)
		}

		t.Run("Should return an InternalError", func(t *testing.T) {
			t.Parallel()

			suite := WhenAuthorizingAnX402QuoteBeforeEach(t)
			initSuite(suite)

			_, err := suite.sut.Authorize(ctx, authorizedQuote(), big.NewInt(10000))

			var internalError *apperrors.InternalError
			assert.ErrorAs(t, err, &internalError)
			assert.Equal(t, "spend_policy_failed", internalError.Code)
		})
	})

	t.Run("Given the token's rule no longer exists", func(t *testing.T) {
		t.Parallel()

		policy := model_builders.NewPaymentSpendPolicyBuilder().Build()
		config := model_builders.NewPaymentSpendRuleX402ConfigBuilder().WithPolicyId(policy.Id).Build()

		initSuite := func(suite *WhenAuthorizingAnX402QuoteTestingSuite) {
			suite.mockPaymentSpendPolicyProvider.EXPECT().Ensure(ctx).Return(policy, nil).Times(1)
			suite.mockUnitOfWorkFactory.EXPECT().Start(ctx).Return(suite.mockUnitOfWork, nil).Times(1)
			suite.mockPaymentSpendRuleX402ConfigRepository.EXPECT().
				GetByPolicyIdAndAsset(suite.unitOfWorkCtx, policy.Id, quoteNetwork, storedQuoteAsset).
				Return(config, nil).
				Times(1)
			suite.mockPaymentSpendRuleRepository.EXPECT().
				GetById(suite.unitOfWorkCtx, config.PaymentSpendRuleId, runtime_common_repositories.ToLock).
				Return(nil, nil).
				Times(1)
		}

		t.Run("Should refuse the quote", func(t *testing.T) {
			t.Parallel()

			suite := WhenAuthorizingAnX402QuoteBeforeEach(t)
			initSuite(suite)

			reservation, err := suite.sut.Authorize(ctx, authorizedQuote(), big.NewInt(10000))

			assert.Nil(t, reservation)
			var unauthorizedError *apperrors.UnauthorizedError
			assert.ErrorAs(t, err, &unauthorizedError)
			assert.Equal(t, "spend_token_has_no_rule", unauthorizedError.Code)
		})
	})

	t.Run("Given the token's rule uses another payment method", func(t *testing.T) {
		t.Parallel()

		policy := model_builders.NewPaymentSpendPolicyBuilder().Build()
		rule := model_builders.NewPaymentSpendRuleBuilder().WithPolicyId(policy.Id).Build()
		rule.Method = enums.PaymentMethod("other")
		config := model_builders.NewPaymentSpendRuleX402ConfigBuilder().
			WithPolicyId(policy.Id).
			WithPaymentSpendRuleId(rule.Id).
			Build()

		initSuite := func(suite *WhenAuthorizingAnX402QuoteTestingSuite) {
			suite.mockPaymentSpendPolicyProvider.EXPECT().Ensure(ctx).Return(policy, nil).Times(1)
			suite.mockUnitOfWorkFactory.EXPECT().Start(ctx).Return(suite.mockUnitOfWork, nil).Times(1)
			suite.mockPaymentSpendRuleX402ConfigRepository.EXPECT().
				GetByPolicyIdAndAsset(suite.unitOfWorkCtx, policy.Id, quoteNetwork, storedQuoteAsset).
				Return(config, nil).
				Times(1)
			suite.mockPaymentSpendRuleRepository.EXPECT().
				GetById(suite.unitOfWorkCtx, rule.Id, runtime_common_repositories.ToLock).
				Return(rule, nil).
				Times(1)
		}

		t.Run("Should refuse the quote", func(t *testing.T) {
			t.Parallel()

			suite := WhenAuthorizingAnX402QuoteBeforeEach(t)
			initSuite(suite)

			reservation, err := suite.sut.Authorize(ctx, authorizedQuote(), big.NewInt(10000))

			assert.Nil(t, reservation)
			var unauthorizedError *apperrors.UnauthorizedError
			assert.ErrorAs(t, err, &unauthorizedError)
			assert.Equal(t, "spend_token_has_no_rule", unauthorizedError.Code)
		})
	})

	t.Run("Given reading the locked token fails", func(t *testing.T) {
		t.Parallel()

		policy := model_builders.NewPaymentSpendPolicyBuilder().Build()
		rule := model_builders.NewPaymentSpendRuleBuilder().WithPolicyId(policy.Id).Build()
		config := model_builders.NewPaymentSpendRuleX402ConfigBuilder().
			WithPolicyId(policy.Id).
			WithPaymentSpendRuleId(rule.Id).
			Build()

		initSuite := func(suite *WhenAuthorizingAnX402QuoteTestingSuite) {
			suite.mockPaymentSpendPolicyProvider.EXPECT().Ensure(ctx).Return(policy, nil).Times(1)
			suite.mockUnitOfWorkFactory.EXPECT().Start(ctx).Return(suite.mockUnitOfWork, nil).Times(1)
			suite.mockPaymentSpendRuleX402ConfigRepository.EXPECT().
				GetByPolicyIdAndAsset(suite.unitOfWorkCtx, policy.Id, quoteNetwork, storedQuoteAsset).
				Return(config, nil).
				Times(1)
			suite.mockPaymentSpendRuleRepository.EXPECT().
				GetById(suite.unitOfWorkCtx, rule.Id, runtime_common_repositories.ToLock).
				Return(rule, nil).
				Times(1)
			suite.mockPaymentSpendRuleX402ConfigRepository.EXPECT().
				GetByRuleId(suite.unitOfWorkCtx, rule.Id, runtime_common_repositories.ToLock).
				Return(nil, assert.AnError).
				Times(1)
		}

		t.Run("Should return an InternalError", func(t *testing.T) {
			t.Parallel()

			suite := WhenAuthorizingAnX402QuoteBeforeEach(t)
			initSuite(suite)

			_, err := suite.sut.Authorize(ctx, authorizedQuote(), big.NewInt(10000))

			var internalError *apperrors.InternalError
			assert.ErrorAs(t, err, &internalError)
			assert.Equal(t, "spend_policy_failed", internalError.Code)
		})
	})

	t.Run("Given the token's config was removed before it was locked", func(t *testing.T) {
		t.Parallel()

		policy := model_builders.NewPaymentSpendPolicyBuilder().Build()
		rule := model_builders.NewPaymentSpendRuleBuilder().WithPolicyId(policy.Id).Build()
		config := model_builders.NewPaymentSpendRuleX402ConfigBuilder().
			WithPolicyId(policy.Id).
			WithPaymentSpendRuleId(rule.Id).
			Build()

		initSuite := func(suite *WhenAuthorizingAnX402QuoteTestingSuite) {
			suite.mockPaymentSpendPolicyProvider.EXPECT().Ensure(ctx).Return(policy, nil).Times(1)
			suite.mockUnitOfWorkFactory.EXPECT().Start(ctx).Return(suite.mockUnitOfWork, nil).Times(1)
			suite.mockPaymentSpendRuleX402ConfigRepository.EXPECT().
				GetByPolicyIdAndAsset(suite.unitOfWorkCtx, policy.Id, quoteNetwork, storedQuoteAsset).
				Return(config, nil).
				Times(1)
			suite.mockPaymentSpendRuleRepository.EXPECT().
				GetById(suite.unitOfWorkCtx, rule.Id, runtime_common_repositories.ToLock).
				Return(rule, nil).
				Times(1)
			suite.mockPaymentSpendRuleX402ConfigRepository.EXPECT().
				GetByRuleId(suite.unitOfWorkCtx, rule.Id, runtime_common_repositories.ToLock).
				Return(nil, nil).
				Times(1)
		}

		t.Run("Should refuse the quote without touching the windows", func(t *testing.T) {
			t.Parallel()

			suite := WhenAuthorizingAnX402QuoteBeforeEach(t)
			initSuite(suite)

			reservation, err := suite.sut.Authorize(ctx, authorizedQuote(), big.NewInt(10000))

			assert.Nil(t, reservation)
			var unauthorizedError *apperrors.UnauthorizedError
			if assert.ErrorAs(t, err, &unauthorizedError) {
				assert.Equal(t, "spend_token_has_no_rule", unauthorizedError.Code)
			}
		})
	})

	t.Run("Given the rule moved to another token before it was locked", func(t *testing.T) {
		t.Parallel()

		policy := model_builders.NewPaymentSpendPolicyBuilder().Build()
		rule := model_builders.NewPaymentSpendRuleBuilder().WithPolicyId(policy.Id).Build()
		config := model_builders.NewPaymentSpendRuleX402ConfigBuilder().
			WithPolicyId(policy.Id).
			WithPaymentSpendRuleId(rule.Id).
			Build()
		moved := model_builders.NewPaymentSpendRuleX402ConfigBuilder().
			WithPolicyId(policy.Id).
			WithPaymentSpendRuleId(rule.Id).
			Build()

		initSuite := func(suite *WhenAuthorizingAnX402QuoteTestingSuite) {
			suite.mockPaymentSpendPolicyProvider.EXPECT().Ensure(ctx).Return(policy, nil).Times(1)
			suite.mockUnitOfWorkFactory.EXPECT().Start(ctx).Return(suite.mockUnitOfWork, nil).Times(1)
			suite.mockPaymentSpendRuleX402ConfigRepository.EXPECT().
				GetByPolicyIdAndAsset(suite.unitOfWorkCtx, policy.Id, quoteNetwork, storedQuoteAsset).
				Return(config, nil).
				Times(1)
			suite.mockPaymentSpendRuleRepository.EXPECT().
				GetById(suite.unitOfWorkCtx, rule.Id, runtime_common_repositories.ToLock).
				Return(rule, nil).
				Times(1)
			suite.mockPaymentSpendRuleX402ConfigRepository.EXPECT().
				GetByRuleId(suite.unitOfWorkCtx, rule.Id, runtime_common_repositories.ToLock).
				Return(moved, nil).
				Times(1)
		}

		t.Run("Should refuse the quote", func(t *testing.T) {
			t.Parallel()

			suite := WhenAuthorizingAnX402QuoteBeforeEach(t)
			initSuite(suite)

			reservation, err := suite.sut.Authorize(ctx, authorizedQuote(), big.NewInt(10000))

			assert.Nil(t, reservation)
			var unauthorizedError *apperrors.UnauthorizedError
			assert.ErrorAs(t, err, &unauthorizedError)
			assert.Equal(t, "spend_token_has_no_rule", unauthorizedError.Code)
		})
	})

	t.Run("Given the token's rule refuses the quote's transfer method", func(t *testing.T) {
		t.Parallel()

		policy := model_builders.NewPaymentSpendPolicyBuilder().Build()
		rule := model_builders.NewPaymentSpendRuleBuilder().WithPolicyId(policy.Id).Build()
		config := model_builders.NewPaymentSpendRuleX402ConfigBuilder().
			WithPolicyId(policy.Id).
			WithPaymentSpendRuleId(rule.Id).
			WithAssetTransferMethod(enums.AssetTransferMethodPermit2).
			Build()

		initSuite := func(suite *WhenAuthorizingAnX402QuoteTestingSuite) {
			suite.mockPaymentSpendPolicyProvider.EXPECT().Ensure(ctx).Return(policy, nil).Times(1)
			suite.mockUnitOfWorkFactory.EXPECT().Start(ctx).Return(suite.mockUnitOfWork, nil).Times(1)
			suite.mockPaymentSpendRuleX402ConfigRepository.EXPECT().
				GetByPolicyIdAndAsset(suite.unitOfWorkCtx, policy.Id, quoteNetwork, storedQuoteAsset).
				Return(config, nil).
				Times(1)
			suite.mockPaymentSpendRuleRepository.EXPECT().
				GetById(suite.unitOfWorkCtx, rule.Id, runtime_common_repositories.ToLock).
				Return(rule, nil).
				Times(1)
			suite.mockPaymentSpendRuleX402ConfigRepository.EXPECT().
				GetByRuleId(suite.unitOfWorkCtx, rule.Id, runtime_common_repositories.ToLock).
				Return(config, nil).
				Times(1)
		}

		t.Run("Should refuse the quote", func(t *testing.T) {
			t.Parallel()

			suite := WhenAuthorizingAnX402QuoteBeforeEach(t)
			initSuite(suite)

			reservation, err := suite.sut.Authorize(ctx, authorizedQuote(), big.NewInt(10000))

			assert.Nil(t, reservation)
			var unauthorizedError *apperrors.UnauthorizedError
			assert.ErrorAs(t, err, &unauthorizedError)
			assert.Equal(t, "spend_transfer_method_not_allowed", unauthorizedError.Code)
		})
	})

	t.Run("Given the per-payment cap is not a usable amount", func(t *testing.T) {
		t.Parallel()

		policy := model_builders.NewPaymentSpendPolicyBuilder().Build()
		rule := model_builders.NewPaymentSpendRuleBuilder().
			WithPolicyId(policy.Id).
			WithMaxAmountPerPayment("not-a-number").
			Build()
		config := model_builders.NewPaymentSpendRuleX402ConfigBuilder().
			WithPolicyId(policy.Id).
			WithPaymentSpendRuleId(rule.Id).
			Build()

		initSuite := func(suite *WhenAuthorizingAnX402QuoteTestingSuite) {
			suite.mockPaymentSpendPolicyProvider.EXPECT().Ensure(ctx).Return(policy, nil).Times(1)
			suite.mockUnitOfWorkFactory.EXPECT().Start(ctx).Return(suite.mockUnitOfWork, nil).Times(1)
			suite.mockPaymentSpendRuleX402ConfigRepository.EXPECT().
				GetByPolicyIdAndAsset(suite.unitOfWorkCtx, policy.Id, quoteNetwork, storedQuoteAsset).
				Return(config, nil).
				Times(1)
			suite.mockPaymentSpendRuleRepository.EXPECT().
				GetById(suite.unitOfWorkCtx, rule.Id, runtime_common_repositories.ToLock).
				Return(rule, nil).
				Times(1)
			suite.mockPaymentSpendRuleX402ConfigRepository.EXPECT().
				GetByRuleId(suite.unitOfWorkCtx, rule.Id, runtime_common_repositories.ToLock).
				Return(config, nil).
				Times(1)
		}

		t.Run("Should return an InternalError", func(t *testing.T) {
			t.Parallel()

			suite := WhenAuthorizingAnX402QuoteBeforeEach(t)
			initSuite(suite)

			_, err := suite.sut.Authorize(ctx, authorizedQuote(), big.NewInt(10000))

			var internalError *apperrors.InternalError
			assert.ErrorAs(t, err, &internalError)
			assert.Equal(t, "spend_policy_failed", internalError.Code)
		})
	})

	t.Run("Given the amount is above the per-payment cap", func(t *testing.T) {
		t.Parallel()

		policy := model_builders.NewPaymentSpendPolicyBuilder().Build()
		rule := model_builders.NewPaymentSpendRuleBuilder().
			WithPolicyId(policy.Id).
			WithMaxAmountPerPayment("9999").
			Build()
		config := model_builders.NewPaymentSpendRuleX402ConfigBuilder().
			WithPolicyId(policy.Id).
			WithPaymentSpendRuleId(rule.Id).
			Build()

		initSuite := func(suite *WhenAuthorizingAnX402QuoteTestingSuite) {
			suite.mockPaymentSpendPolicyProvider.EXPECT().Ensure(ctx).Return(policy, nil).Times(1)
			suite.mockUnitOfWorkFactory.EXPECT().Start(ctx).Return(suite.mockUnitOfWork, nil).Times(1)
			suite.mockPaymentSpendRuleX402ConfigRepository.EXPECT().
				GetByPolicyIdAndAsset(suite.unitOfWorkCtx, policy.Id, quoteNetwork, storedQuoteAsset).
				Return(config, nil).
				Times(1)
			suite.mockPaymentSpendRuleRepository.EXPECT().
				GetById(suite.unitOfWorkCtx, rule.Id, runtime_common_repositories.ToLock).
				Return(rule, nil).
				Times(1)
			suite.mockPaymentSpendRuleX402ConfigRepository.EXPECT().
				GetByRuleId(suite.unitOfWorkCtx, rule.Id, runtime_common_repositories.ToLock).
				Return(config, nil).
				Times(1)
		}

		t.Run("Should refuse the quote", func(t *testing.T) {
			t.Parallel()

			suite := WhenAuthorizingAnX402QuoteBeforeEach(t)
			initSuite(suite)

			reservation, err := suite.sut.Authorize(ctx, authorizedQuote(), big.NewInt(10000))

			assert.Nil(t, reservation)
			var unauthorizedError *apperrors.UnauthorizedError
			assert.ErrorAs(t, err, &unauthorizedError)
			assert.Equal(t, "spend_cap_exceeded", unauthorizedError.Code)
		})
	})

	t.Run("Given reading the rule's windows fails", func(t *testing.T) {
		t.Parallel()

		policy := model_builders.NewPaymentSpendPolicyBuilder().Build()
		rule := model_builders.NewPaymentSpendRuleBuilder().WithPolicyId(policy.Id).Build()
		config := model_builders.NewPaymentSpendRuleX402ConfigBuilder().
			WithPolicyId(policy.Id).
			WithPaymentSpendRuleId(rule.Id).
			Build()

		initSuite := func(suite *WhenAuthorizingAnX402QuoteTestingSuite) {
			suite.mockPaymentSpendPolicyProvider.EXPECT().Ensure(ctx).Return(policy, nil).Times(1)
			suite.mockUnitOfWorkFactory.EXPECT().Start(ctx).Return(suite.mockUnitOfWork, nil).Times(1)
			suite.mockPaymentSpendRuleX402ConfigRepository.EXPECT().
				GetByPolicyIdAndAsset(suite.unitOfWorkCtx, policy.Id, quoteNetwork, storedQuoteAsset).
				Return(config, nil).
				Times(1)
			suite.mockPaymentSpendRuleRepository.EXPECT().
				GetById(suite.unitOfWorkCtx, rule.Id, runtime_common_repositories.ToLock).
				Return(rule, nil).
				Times(1)
			suite.mockPaymentSpendRuleX402ConfigRepository.EXPECT().
				GetByRuleId(suite.unitOfWorkCtx, rule.Id, runtime_common_repositories.ToLock).
				Return(config, nil).
				Times(1)
			suite.mockPaymentSpendRuleWindowRepository.EXPECT().
				ListByRuleId(suite.unitOfWorkCtx, rule.Id, runtime_common_repositories.ToLock).
				Return(nil, assert.AnError).
				Times(1)
		}

		t.Run("Should return an InternalError", func(t *testing.T) {
			t.Parallel()

			suite := WhenAuthorizingAnX402QuoteBeforeEach(t)
			initSuite(suite)

			_, err := suite.sut.Authorize(ctx, authorizedQuote(), big.NewInt(10000))

			var internalError *apperrors.InternalError
			assert.ErrorAs(t, err, &internalError)
			assert.Equal(t, "spend_policy_failed", internalError.Code)
		})
	})

	t.Run("Given a window budget is not a usable amount", func(t *testing.T) {
		t.Parallel()

		policy := model_builders.NewPaymentSpendPolicyBuilder().Build()
		rule := model_builders.NewPaymentSpendRuleBuilder().WithPolicyId(policy.Id).Build()
		config := model_builders.NewPaymentSpendRuleX402ConfigBuilder().
			WithPolicyId(policy.Id).
			WithPaymentSpendRuleId(rule.Id).
			Build()
		window := model_builders.NewPaymentSpendRuleWindowBuilder().
			WithPaymentSpendRuleId(rule.Id).
			WithMaxAmount("not-a-number").
			Build()

		initSuite := func(suite *WhenAuthorizingAnX402QuoteTestingSuite) {
			suite.mockPaymentSpendPolicyProvider.EXPECT().Ensure(ctx).Return(policy, nil).Times(1)
			suite.mockUnitOfWorkFactory.EXPECT().Start(ctx).Return(suite.mockUnitOfWork, nil).Times(1)
			suite.mockPaymentSpendRuleX402ConfigRepository.EXPECT().
				GetByPolicyIdAndAsset(suite.unitOfWorkCtx, policy.Id, quoteNetwork, storedQuoteAsset).
				Return(config, nil).
				Times(1)
			suite.mockPaymentSpendRuleRepository.EXPECT().
				GetById(suite.unitOfWorkCtx, rule.Id, runtime_common_repositories.ToLock).
				Return(rule, nil).
				Times(1)
			suite.mockPaymentSpendRuleX402ConfigRepository.EXPECT().
				GetByRuleId(suite.unitOfWorkCtx, rule.Id, runtime_common_repositories.ToLock).
				Return(config, nil).
				Times(1)
			suite.mockPaymentSpendRuleWindowRepository.EXPECT().
				ListByRuleId(suite.unitOfWorkCtx, rule.Id, runtime_common_repositories.ToLock).
				Return([]models.PaymentSpendRuleWindow{*window}, nil).
				Times(1)
		}

		t.Run("Should return an InternalError", func(t *testing.T) {
			t.Parallel()

			suite := WhenAuthorizingAnX402QuoteBeforeEach(t)
			initSuite(suite)

			_, err := suite.sut.Authorize(ctx, authorizedQuote(), big.NewInt(10000))

			var internalError *apperrors.InternalError
			assert.ErrorAs(t, err, &internalError)
			assert.Equal(t, "spend_policy_failed", internalError.Code)
		})
	})

	t.Run("Given one window cannot absorb the amount", func(t *testing.T) {
		t.Parallel()

		policy := model_builders.NewPaymentSpendPolicyBuilder().Build()
		rule := model_builders.NewPaymentSpendRuleBuilder().WithPolicyId(policy.Id).WithoutMaxAmountPerPayment().Build()
		config := model_builders.NewPaymentSpendRuleX402ConfigBuilder().
			WithPolicyId(policy.Id).
			WithPaymentSpendRuleId(rule.Id).
			Build()
		roomyWindow := model_builders.NewPaymentSpendRuleWindowBuilder().
			WithPaymentSpendRuleId(rule.Id).
			WithMaxAmount("1000000").
			Build()
		tightWindow := model_builders.NewPaymentSpendRuleWindowBuilder().
			WithPaymentSpendRuleId(rule.Id).
			WithWindowSeconds(604800).
			WithMaxAmount("9999").
			Build()

		// No Update or Commit expectation: a refusal must record no spend on any window.
		initSuite := func(suite *WhenAuthorizingAnX402QuoteTestingSuite) {
			suite.mockPaymentSpendPolicyProvider.EXPECT().Ensure(ctx).Return(policy, nil).Times(1)
			suite.mockUnitOfWorkFactory.EXPECT().Start(ctx).Return(suite.mockUnitOfWork, nil).Times(1)
			suite.mockPaymentSpendRuleX402ConfigRepository.EXPECT().
				GetByPolicyIdAndAsset(suite.unitOfWorkCtx, policy.Id, quoteNetwork, storedQuoteAsset).
				Return(config, nil).
				Times(1)
			suite.mockPaymentSpendRuleRepository.EXPECT().
				GetById(suite.unitOfWorkCtx, rule.Id, runtime_common_repositories.ToLock).
				Return(rule, nil).
				Times(1)
			suite.mockPaymentSpendRuleX402ConfigRepository.EXPECT().
				GetByRuleId(suite.unitOfWorkCtx, rule.Id, runtime_common_repositories.ToLock).
				Return(config, nil).
				Times(1)
			suite.mockPaymentSpendRuleWindowRepository.EXPECT().
				ListByRuleId(suite.unitOfWorkCtx, rule.Id, runtime_common_repositories.ToLock).
				Return([]models.PaymentSpendRuleWindow{*roomyWindow, *tightWindow}, nil).
				Times(1)
		}

		t.Run("Should refuse the quote", func(t *testing.T) {
			t.Parallel()

			suite := WhenAuthorizingAnX402QuoteBeforeEach(t)
			initSuite(suite)

			reservation, err := suite.sut.Authorize(ctx, authorizedQuote(), big.NewInt(10000))

			assert.Nil(t, reservation)
			var unauthorizedError *apperrors.UnauthorizedError
			assert.ErrorAs(t, err, &unauthorizedError)
			assert.Equal(t, "spend_window_exhausted", unauthorizedError.Code)
		})
	})

	t.Run("Given the running period has too little left for the amount", func(t *testing.T) {
		t.Parallel()

		policy := model_builders.NewPaymentSpendPolicyBuilder().Build()
		rule := model_builders.NewPaymentSpendRuleBuilder().WithPolicyId(policy.Id).WithoutMaxAmountPerPayment().Build()
		config := model_builders.NewPaymentSpendRuleX402ConfigBuilder().
			WithPolicyId(policy.Id).
			WithPaymentSpendRuleId(rule.Id).
			Build()
		window := model_builders.NewPaymentSpendRuleWindowBuilder().
			WithPaymentSpendRuleId(rule.Id).
			WithMaxAmount("1000000").
			WithSpent("995000").
			WithStartedAt(time.Now().Add(-time.Hour)).
			Build()

		// No Update or Commit expectation: a refusal must record no spend on any window.
		initSuite := func(suite *WhenAuthorizingAnX402QuoteTestingSuite) {
			suite.mockPaymentSpendPolicyProvider.EXPECT().Ensure(ctx).Return(policy, nil).Times(1)
			suite.mockUnitOfWorkFactory.EXPECT().Start(ctx).Return(suite.mockUnitOfWork, nil).Times(1)
			suite.mockPaymentSpendRuleX402ConfigRepository.EXPECT().
				GetByPolicyIdAndAsset(suite.unitOfWorkCtx, policy.Id, quoteNetwork, storedQuoteAsset).
				Return(config, nil).
				Times(1)
			suite.mockPaymentSpendRuleRepository.EXPECT().
				GetById(suite.unitOfWorkCtx, rule.Id, runtime_common_repositories.ToLock).
				Return(rule, nil).
				Times(1)
			suite.mockPaymentSpendRuleX402ConfigRepository.EXPECT().
				GetByRuleId(suite.unitOfWorkCtx, rule.Id, runtime_common_repositories.ToLock).
				Return(config, nil).
				Times(1)
			suite.mockPaymentSpendRuleWindowRepository.EXPECT().
				ListByRuleId(suite.unitOfWorkCtx, rule.Id, runtime_common_repositories.ToLock).
				Return([]models.PaymentSpendRuleWindow{*window}, nil).
				Times(1)
		}

		t.Run("Should refuse the quote", func(t *testing.T) {
			t.Parallel()

			suite := WhenAuthorizingAnX402QuoteBeforeEach(t)
			initSuite(suite)

			reservation, err := suite.sut.Authorize(ctx, authorizedQuote(), big.NewInt(10000))

			assert.Nil(t, reservation)
			var unauthorizedError *apperrors.UnauthorizedError
			assert.ErrorAs(t, err, &unauthorizedError)
			assert.Equal(t, "spend_window_exhausted", unauthorizedError.Code)
		})
	})

	t.Run("Given saving a window fails", func(t *testing.T) {
		t.Parallel()

		policy := model_builders.NewPaymentSpendPolicyBuilder().Build()
		rule := model_builders.NewPaymentSpendRuleBuilder().WithPolicyId(policy.Id).Build()
		config := model_builders.NewPaymentSpendRuleX402ConfigBuilder().
			WithPolicyId(policy.Id).
			WithPaymentSpendRuleId(rule.Id).
			Build()
		window := model_builders.NewPaymentSpendRuleWindowBuilder().WithPaymentSpendRuleId(rule.Id).Build()

		initSuite := func(suite *WhenAuthorizingAnX402QuoteTestingSuite) {
			suite.mockPaymentSpendPolicyProvider.EXPECT().Ensure(ctx).Return(policy, nil).Times(1)
			suite.mockUnitOfWorkFactory.EXPECT().Start(ctx).Return(suite.mockUnitOfWork, nil).Times(1)
			suite.mockPaymentSpendRuleX402ConfigRepository.EXPECT().
				GetByPolicyIdAndAsset(suite.unitOfWorkCtx, policy.Id, quoteNetwork, storedQuoteAsset).
				Return(config, nil).
				Times(1)
			suite.mockPaymentSpendRuleRepository.EXPECT().
				GetById(suite.unitOfWorkCtx, rule.Id, runtime_common_repositories.ToLock).
				Return(rule, nil).
				Times(1)
			suite.mockPaymentSpendRuleX402ConfigRepository.EXPECT().
				GetByRuleId(suite.unitOfWorkCtx, rule.Id, runtime_common_repositories.ToLock).
				Return(config, nil).
				Times(1)
			suite.mockPaymentSpendRuleWindowRepository.EXPECT().
				ListByRuleId(suite.unitOfWorkCtx, rule.Id, runtime_common_repositories.ToLock).
				Return([]models.PaymentSpendRuleWindow{*window}, nil).
				Times(1)
			suite.mockPaymentSpendRuleWindowRepository.EXPECT().
				Update(suite.unitOfWorkCtx, gomock.Any()).
				Return(nil, assert.AnError).
				Times(1)
		}

		t.Run("Should return an InternalError", func(t *testing.T) {
			t.Parallel()

			suite := WhenAuthorizingAnX402QuoteBeforeEach(t)
			initSuite(suite)

			_, err := suite.sut.Authorize(ctx, authorizedQuote(), big.NewInt(10000))

			var internalError *apperrors.InternalError
			assert.ErrorAs(t, err, &internalError)
			assert.Equal(t, "spend_policy_failed", internalError.Code)
		})
	})

	t.Run("Given committing the spend fails", func(t *testing.T) {
		t.Parallel()

		policy := model_builders.NewPaymentSpendPolicyBuilder().Build()
		rule := model_builders.NewPaymentSpendRuleBuilder().WithPolicyId(policy.Id).Build()
		config := model_builders.NewPaymentSpendRuleX402ConfigBuilder().
			WithPolicyId(policy.Id).
			WithPaymentSpendRuleId(rule.Id).
			Build()

		initSuite := func(suite *WhenAuthorizingAnX402QuoteTestingSuite) {
			suite.mockPaymentSpendPolicyProvider.EXPECT().Ensure(ctx).Return(policy, nil).Times(1)
			suite.mockUnitOfWorkFactory.EXPECT().Start(ctx).Return(suite.mockUnitOfWork, nil).Times(1)
			suite.mockPaymentSpendRuleX402ConfigRepository.EXPECT().
				GetByPolicyIdAndAsset(suite.unitOfWorkCtx, policy.Id, quoteNetwork, storedQuoteAsset).
				Return(config, nil).
				Times(1)
			suite.mockPaymentSpendRuleRepository.EXPECT().
				GetById(suite.unitOfWorkCtx, rule.Id, runtime_common_repositories.ToLock).
				Return(rule, nil).
				Times(1)
			suite.mockPaymentSpendRuleX402ConfigRepository.EXPECT().
				GetByRuleId(suite.unitOfWorkCtx, rule.Id, runtime_common_repositories.ToLock).
				Return(config, nil).
				Times(1)
			suite.mockPaymentSpendRuleWindowRepository.EXPECT().
				ListByRuleId(suite.unitOfWorkCtx, rule.Id, runtime_common_repositories.ToLock).
				Return([]models.PaymentSpendRuleWindow{}, nil).
				Times(1)
			suite.mockUnitOfWork.EXPECT().Commit().Return(assert.AnError).Times(1)
		}

		t.Run("Should return an InternalError", func(t *testing.T) {
			t.Parallel()

			suite := WhenAuthorizingAnX402QuoteBeforeEach(t)
			initSuite(suite)

			_, err := suite.sut.Authorize(ctx, authorizedQuote(), big.NewInt(10000))

			var internalError *apperrors.InternalError
			assert.ErrorAs(t, err, &internalError)
			assert.Equal(t, "spend_policy_failed", internalError.Code)
		})
	})

	t.Run("Given the token's rule has a cap and no window", func(t *testing.T) {
		t.Parallel()

		policy := model_builders.NewPaymentSpendPolicyBuilder().Build()
		rule := model_builders.NewPaymentSpendRuleBuilder().
			WithPolicyId(policy.Id).
			WithMaxAmountPerPayment("10000").
			Build()
		config := model_builders.NewPaymentSpendRuleX402ConfigBuilder().
			WithPolicyId(policy.Id).
			WithPaymentSpendRuleId(rule.Id).
			Build()

		initSuite := func(suite *WhenAuthorizingAnX402QuoteTestingSuite) {
			suite.mockPaymentSpendPolicyProvider.EXPECT().Ensure(ctx).Return(policy, nil).Times(1)
			suite.mockUnitOfWorkFactory.EXPECT().Start(ctx).Return(suite.mockUnitOfWork, nil).Times(1)
			suite.mockPaymentSpendRuleX402ConfigRepository.EXPECT().
				GetByPolicyIdAndAsset(suite.unitOfWorkCtx, policy.Id, quoteNetwork, storedQuoteAsset).
				Return(config, nil).
				Times(1)
			suite.mockPaymentSpendRuleRepository.EXPECT().
				GetById(suite.unitOfWorkCtx, rule.Id, runtime_common_repositories.ToLock).
				Return(rule, nil).
				Times(1)
			suite.mockPaymentSpendRuleX402ConfigRepository.EXPECT().
				GetByRuleId(suite.unitOfWorkCtx, rule.Id, runtime_common_repositories.ToLock).
				Return(config, nil).
				Times(1)
			suite.mockPaymentSpendRuleWindowRepository.EXPECT().
				ListByRuleId(suite.unitOfWorkCtx, rule.Id, runtime_common_repositories.ToLock).
				Return([]models.PaymentSpendRuleWindow{}, nil).
				Times(1)
			suite.mockUnitOfWork.EXPECT().Commit().Return(nil).Times(1)
		}

		// The cap is inclusive: an amount equal to it is still within the rule.
		t.Run("Should authorize an amount equal to the cap", func(t *testing.T) {
			t.Parallel()

			suite := WhenAuthorizingAnX402QuoteBeforeEach(t)
			initSuite(suite)

			reservation, err := suite.sut.Authorize(ctx, authorizedQuote(), big.NewInt(10000))

			assert.NoError(t, err)
			assert.NotNil(t, reservation)
		})
	})

	t.Run("Given the amount fits the cap and every window", func(t *testing.T) {
		t.Parallel()

		policy := model_builders.NewPaymentSpendPolicyBuilder().Build()
		rule := model_builders.NewPaymentSpendRuleBuilder().WithPolicyId(policy.Id).Build()
		config := model_builders.NewPaymentSpendRuleX402ConfigBuilder().
			WithPolicyId(policy.Id).
			WithPaymentSpendRuleId(rule.Id).
			Build()
		dailyWindow := model_builders.NewPaymentSpendRuleWindowBuilder().WithPaymentSpendRuleId(rule.Id).Build()
		weeklyWindow := model_builders.NewPaymentSpendRuleWindowBuilder().
			WithPaymentSpendRuleId(rule.Id).
			WithWindowSeconds(604800).
			Build()

		initSuite := func(suite *WhenAuthorizingAnX402QuoteTestingSuite) {
			suite.mockPaymentSpendPolicyProvider.EXPECT().Ensure(ctx).Return(policy, nil).Times(1)
			suite.mockUnitOfWorkFactory.EXPECT().Start(ctx).Return(suite.mockUnitOfWork, nil).Times(1)
			suite.mockPaymentSpendRuleX402ConfigRepository.EXPECT().
				GetByPolicyIdAndAsset(suite.unitOfWorkCtx, policy.Id, quoteNetwork, storedQuoteAsset).
				Return(config, nil).
				Times(1)
			suite.mockPaymentSpendRuleRepository.EXPECT().
				GetById(suite.unitOfWorkCtx, rule.Id, runtime_common_repositories.ToLock).
				Return(rule, nil).
				Times(1)
			suite.mockPaymentSpendRuleX402ConfigRepository.EXPECT().
				GetByRuleId(suite.unitOfWorkCtx, rule.Id, runtime_common_repositories.ToLock).
				Return(config, nil).
				Times(1)
			suite.mockPaymentSpendRuleWindowRepository.EXPECT().
				ListByRuleId(suite.unitOfWorkCtx, rule.Id, runtime_common_repositories.ToLock).
				Return([]models.PaymentSpendRuleWindow{*dailyWindow, *weeklyWindow}, nil).
				Times(1)
		}

		t.Run("Should authorize the quote", func(t *testing.T) {
			t.Parallel()

			suite := WhenAuthorizingAnX402QuoteBeforeEach(t)
			initSuite(suite)
			suite.mockPaymentSpendRuleWindowRepository.EXPECT().
				Update(suite.unitOfWorkCtx, gomock.Any()).
				Return(nil, nil).
				Times(2)
			suite.mockUnitOfWork.EXPECT().Commit().Return(nil).Times(1)

			reservation, err := suite.sut.Authorize(ctx, authorizedQuote(), big.NewInt(10000))

			assert.NoError(t, err)
			assert.NotNil(t, reservation)
		})

		t.Run("Should return a reservation of the amount on each window's period", func(t *testing.T) {
			t.Parallel()

			suite := WhenAuthorizingAnX402QuoteBeforeEach(t)
			initSuite(suite)
			suite.mockPaymentSpendRuleWindowRepository.EXPECT().
				Update(suite.unitOfWorkCtx, gomock.Any()).
				Return(nil, nil).
				Times(2)
			suite.mockUnitOfWork.EXPECT().Commit().Return(nil).Times(1)

			reservation, err := suite.sut.Authorize(ctx, authorizedQuote(), big.NewInt(10000))

			assert.NoError(t, err)
			if assert.NotNil(t, reservation) {
				assert.Equal(t, rule.Id, reservation.RuleId)
				assert.Equal(t, big.NewInt(10000), reservation.Amount)
				assert.Len(t, reservation.WindowStarts, 2)
				assert.Contains(t, reservation.WindowStarts, dailyWindow.Id)
				assert.Contains(t, reservation.WindowStarts, weeklyWindow.Id)
				// Whole seconds, so Release still matches the period after any driver's datetime round trip.
				dailyStartedAt := reservation.WindowStarts[dailyWindow.Id]
				assert.Equal(t, dailyStartedAt.Truncate(time.Second), dailyStartedAt)
				weeklyStartedAt := reservation.WindowStarts[weeklyWindow.Id]
				assert.Equal(t, weeklyStartedAt.Truncate(time.Second), weeklyStartedAt)
			}
		})

		// Recording on every window keeps each budget its own, not just the binding one.
		t.Run("Should record the amount on every window", func(t *testing.T) {
			t.Parallel()

			suite := WhenAuthorizingAnX402QuoteBeforeEach(t)
			initSuite(suite)
			suite.mockPaymentSpendRuleWindowRepository.EXPECT().
				Update(suite.unitOfWorkCtx, windowWithSpend{id: dailyWindow.Id.String(), spent: "10000"}).
				Return(nil, nil).
				Times(1)
			suite.mockPaymentSpendRuleWindowRepository.EXPECT().
				Update(suite.unitOfWorkCtx, windowWithSpend{id: weeklyWindow.Id.String(), spent: "10000"}).
				Return(nil, nil).
				Times(1)
			suite.mockUnitOfWork.EXPECT().Commit().Return(nil).Times(1)

			_, err := suite.sut.Authorize(ctx, authorizedQuote(), big.NewInt(10000))

			assert.NoError(t, err)
		})
	})

	t.Run("Given the running period has room left", func(t *testing.T) {
		t.Parallel()

		startedAt := time.Now().Add(-time.Hour).Truncate(time.Second)
		policy := model_builders.NewPaymentSpendPolicyBuilder().Build()
		rule := model_builders.NewPaymentSpendRuleBuilder().WithPolicyId(policy.Id).WithoutMaxAmountPerPayment().Build()
		config := model_builders.NewPaymentSpendRuleX402ConfigBuilder().
			WithPolicyId(policy.Id).
			WithPaymentSpendRuleId(rule.Id).
			Build()
		window := model_builders.NewPaymentSpendRuleWindowBuilder().
			WithPaymentSpendRuleId(rule.Id).
			WithMaxAmount("1000000").
			WithSpent("5000").
			WithStartedAt(startedAt).
			Build()

		initSuite := func(suite *WhenAuthorizingAnX402QuoteTestingSuite) {
			suite.mockPaymentSpendPolicyProvider.EXPECT().Ensure(ctx).Return(policy, nil).Times(1)
			suite.mockUnitOfWorkFactory.EXPECT().Start(ctx).Return(suite.mockUnitOfWork, nil).Times(1)
			suite.mockPaymentSpendRuleX402ConfigRepository.EXPECT().
				GetByPolicyIdAndAsset(suite.unitOfWorkCtx, policy.Id, quoteNetwork, storedQuoteAsset).
				Return(config, nil).
				Times(1)
			suite.mockPaymentSpendRuleRepository.EXPECT().
				GetById(suite.unitOfWorkCtx, rule.Id, runtime_common_repositories.ToLock).
				Return(rule, nil).
				Times(1)
			suite.mockPaymentSpendRuleX402ConfigRepository.EXPECT().
				GetByRuleId(suite.unitOfWorkCtx, rule.Id, runtime_common_repositories.ToLock).
				Return(config, nil).
				Times(1)
			suite.mockPaymentSpendRuleWindowRepository.EXPECT().
				ListByRuleId(suite.unitOfWorkCtx, rule.Id, runtime_common_repositories.ToLock).
				Return([]models.PaymentSpendRuleWindow{*window}, nil).
				Times(1)
		}

		t.Run("Should add the amount to what the period has absorbed", func(t *testing.T) {
			t.Parallel()

			suite := WhenAuthorizingAnX402QuoteBeforeEach(t)
			initSuite(suite)
			suite.mockPaymentSpendRuleWindowRepository.EXPECT().
				Update(suite.unitOfWorkCtx, windowWithSpend{id: window.Id.String(), spent: "15000"}).
				Return(nil, nil).
				Times(1)
			suite.mockUnitOfWork.EXPECT().Commit().Return(nil).Times(1)

			_, err := suite.sut.Authorize(ctx, authorizedQuote(), big.NewInt(10000))

			assert.NoError(t, err)
		})

		t.Run("Should reserve the amount on the running period", func(t *testing.T) {
			t.Parallel()

			suite := WhenAuthorizingAnX402QuoteBeforeEach(t)
			initSuite(suite)
			suite.mockPaymentSpendRuleWindowRepository.EXPECT().
				Update(suite.unitOfWorkCtx, gomock.Any()).
				Return(nil, nil).
				Times(1)
			suite.mockUnitOfWork.EXPECT().Commit().Return(nil).Times(1)

			reservation, err := suite.sut.Authorize(ctx, authorizedQuote(), big.NewInt(10000))

			assert.NoError(t, err)
			if assert.NotNil(t, reservation) {
				assert.Equal(t, startedAt, reservation.WindowStarts[window.Id])
			}
		})
	})

	t.Run("Given the period has elapsed", func(t *testing.T) {
		t.Parallel()

		startedAt := time.Now().Add(-25 * time.Hour).Truncate(time.Second)
		policy := model_builders.NewPaymentSpendPolicyBuilder().Build()
		rule := model_builders.NewPaymentSpendRuleBuilder().WithPolicyId(policy.Id).WithoutMaxAmountPerPayment().Build()
		config := model_builders.NewPaymentSpendRuleX402ConfigBuilder().
			WithPolicyId(policy.Id).
			WithPaymentSpendRuleId(rule.Id).
			Build()
		window := model_builders.NewPaymentSpendRuleWindowBuilder().
			WithPaymentSpendRuleId(rule.Id).
			WithMaxAmount("1000000").
			WithSpent("1000000").
			WithStartedAt(startedAt).
			Build()

		initSuite := func(suite *WhenAuthorizingAnX402QuoteTestingSuite) {
			suite.mockPaymentSpendPolicyProvider.EXPECT().Ensure(ctx).Return(policy, nil).Times(1)
			suite.mockUnitOfWorkFactory.EXPECT().Start(ctx).Return(suite.mockUnitOfWork, nil).Times(1)
			suite.mockPaymentSpendRuleX402ConfigRepository.EXPECT().
				GetByPolicyIdAndAsset(suite.unitOfWorkCtx, policy.Id, quoteNetwork, storedQuoteAsset).
				Return(config, nil).
				Times(1)
			suite.mockPaymentSpendRuleRepository.EXPECT().
				GetById(suite.unitOfWorkCtx, rule.Id, runtime_common_repositories.ToLock).
				Return(rule, nil).
				Times(1)
			suite.mockPaymentSpendRuleX402ConfigRepository.EXPECT().
				GetByRuleId(suite.unitOfWorkCtx, rule.Id, runtime_common_repositories.ToLock).
				Return(config, nil).
				Times(1)
			suite.mockPaymentSpendRuleWindowRepository.EXPECT().
				ListByRuleId(suite.unitOfWorkCtx, rule.Id, runtime_common_repositories.ToLock).
				Return([]models.PaymentSpendRuleWindow{*window}, nil).
				Times(1)
		}

		// The budget returns on its own; nothing needs resetting by hand.
		t.Run("Should measure the amount against a fresh period", func(t *testing.T) {
			t.Parallel()

			suite := WhenAuthorizingAnX402QuoteBeforeEach(t)
			initSuite(suite)
			suite.mockPaymentSpendRuleWindowRepository.EXPECT().
				Update(suite.unitOfWorkCtx, windowWithSpend{id: window.Id.String(), spent: "10000"}).
				Return(nil, nil).
				Times(1)
			suite.mockUnitOfWork.EXPECT().Commit().Return(nil).Times(1)

			_, err := suite.sut.Authorize(ctx, authorizedQuote(), big.NewInt(10000))

			assert.NoError(t, err)
		})

		t.Run("Should reserve the amount on the new period", func(t *testing.T) {
			t.Parallel()

			suite := WhenAuthorizingAnX402QuoteBeforeEach(t)
			initSuite(suite)
			suite.mockPaymentSpendRuleWindowRepository.EXPECT().
				Update(suite.unitOfWorkCtx, gomock.Any()).
				Return(nil, nil).
				Times(1)
			suite.mockUnitOfWork.EXPECT().Commit().Return(nil).Times(1)

			reservation, err := suite.sut.Authorize(ctx, authorizedQuote(), big.NewInt(10000))

			assert.NoError(t, err)
			if assert.NotNil(t, reservation) {
				assert.True(t, reservation.WindowStarts[window.Id].After(startedAt))
			}
		})
	})
}

// windowWithSpend matches the window with that id once the spend has been recorded on it.
type windowWithSpend struct {
	id    string
	spent string
}

func (m windowWithSpend) Matches(x any) bool {
	window, ok := x.(*models.PaymentSpendRuleWindow)
	return ok && window.Id.String() == m.id && window.Spent == m.spent
}

func (m windowWithSpend) String() string {
	return "window " + m.id + " with spent " + m.spent
}
