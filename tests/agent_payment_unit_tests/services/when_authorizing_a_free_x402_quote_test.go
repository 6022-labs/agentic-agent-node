package services_test

import (
	"context"
	"testing"

	"github.com/6022-labs/agentic-agent-node/src/agent_payment/services"
	apperrors "github.com/6022-labs/agentic-agent-node/src/common/errors"
	payment_repositories_mocks "github.com/6022-labs/agentic-agent-node/tests/agent_payment_mocks/repositories_mocks"
	payment_interfaces_mocks "github.com/6022-labs/agentic-agent-node/tests/agent_payment_mocks/services_mocks/interfaces_mocks"
	"github.com/6022-labs/agentic-agent-node/tests/common_tests/model_builders"
	runtime_common_mocks "github.com/6022-labs/agentic-agent-node/tests/runtime_common_mocks/services_mocks/interfaces_mocks"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

type WhenAuthorizingAFreeX402QuoteTestingSuite struct {
	sut *services.X402PaymentSpendPolicyEvaluator

	mockPaymentSpendPolicyProvider *payment_interfaces_mocks.MockPaymentSpendPolicyProviderInterface
}

func WhenAuthorizingAFreeX402QuoteBeforeEach(t *testing.T) *WhenAuthorizingAFreeX402QuoteTestingSuite {
	mockController := gomock.NewController(t)

	mockPaymentSpendPolicyProvider := payment_interfaces_mocks.NewMockPaymentSpendPolicyProviderInterface(
		mockController,
	)

	return &WhenAuthorizingAFreeX402QuoteTestingSuite{
		sut: services.NewX402PaymentSpendPolicyEvaluator(
			mockPaymentSpendPolicyProvider,
			payment_repositories_mocks.NewMockPaymentSpendRuleRepositoryInterface(mockController),
			payment_repositories_mocks.NewMockPaymentSpendRuleX402ConfigRepositoryInterface(mockController),
			payment_repositories_mocks.NewMockPaymentSpendRuleWindowRepositoryInterface(mockController),
			runtime_common_mocks.NewMockUnitOfWorkFactoryInterface(mockController),
		),
		mockPaymentSpendPolicyProvider: mockPaymentSpendPolicyProvider,
	}
}

func TestWhenAuthorizingAFreeX402Quote(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	t.Run("Given ensuring the policy fails", func(t *testing.T) {
		t.Parallel()

		initSuite := func(suite *WhenAuthorizingAFreeX402QuoteTestingSuite) {
			suite.mockPaymentSpendPolicyProvider.EXPECT().
				Ensure(ctx).
				Return(nil, apperrors.NewInternalError("spend_policy_failed", nil)).
				Times(1)
		}

		t.Run("Should return the error", func(t *testing.T) {
			t.Parallel()

			suite := WhenAuthorizingAFreeX402QuoteBeforeEach(t)
			initSuite(suite)

			_, err := suite.sut.AuthorizeFree(ctx)

			var internalError *apperrors.InternalError
			assert.ErrorAs(t, err, &internalError)
			assert.Equal(t, "spend_policy_failed", internalError.Code)
		})
	})

	t.Run("Given the agent is not minted", func(t *testing.T) {
		t.Parallel()

		initSuite := func(suite *WhenAuthorizingAFreeX402QuoteTestingSuite) {
			suite.mockPaymentSpendPolicyProvider.EXPECT().
				Ensure(ctx).
				Return(nil, apperrors.NewConflictError("agent_not_minted", "")).
				Times(1)
		}

		t.Run("Should refuse the quote", func(t *testing.T) {
			t.Parallel()

			suite := WhenAuthorizingAFreeX402QuoteBeforeEach(t)
			initSuite(suite)

			reservation, err := suite.sut.AuthorizeFree(ctx)

			assert.Nil(t, reservation)
			var unauthorizedError *apperrors.UnauthorizedError
			assert.ErrorAs(t, err, &unauthorizedError)
			assert.Equal(t, "spend_agent_not_minted", unauthorizedError.Code)
		})
	})

	t.Run("Given a minted agent whose policy pays nobody", func(t *testing.T) {
		t.Parallel()

		policy := model_builders.NewPaymentSpendPolicyBuilder().WithPayToPatterns([]string{}).Build()

		initSuite := func(suite *WhenAuthorizingAFreeX402QuoteTestingSuite) {
			suite.mockPaymentSpendPolicyProvider.EXPECT().Ensure(ctx).Return(policy, nil).Times(1)
		}

		t.Run("Should reserve nothing against any rule", func(t *testing.T) {
			t.Parallel()

			suite := WhenAuthorizingAFreeX402QuoteBeforeEach(t)
			initSuite(suite)

			reservation, err := suite.sut.AuthorizeFree(ctx)

			assert.NoError(t, err)
			assert.Equal(t, uuid.Nil, reservation.RuleId)
			assert.Equal(t, 0, reservation.Amount.Sign())
			assert.Empty(t, reservation.WindowStarts)
		})
	})
}
