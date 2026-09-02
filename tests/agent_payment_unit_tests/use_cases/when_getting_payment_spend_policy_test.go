package use_cases_test

import (
	"context"
	"testing"

	"github.com/6022-labs/agentic-agent-node/src/agent_payment/use_cases"
	apperrors "github.com/6022-labs/agentic-agent-node/src/common/errors"
	payment_repositories_mocks "github.com/6022-labs/agentic-agent-node/tests/agent_payment_mocks/repositories_mocks"
	payment_interfaces_mocks "github.com/6022-labs/agentic-agent-node/tests/agent_payment_mocks/services_mocks/interfaces_mocks"
	"github.com/6022-labs/agentic-agent-node/tests/common_tests/model_builders"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

type WhenGettingPaymentSpendPolicyTestingSuite struct {
	sut *use_cases.GetPaymentSpendPolicy

	mockPaymentSpendPolicyProvider *payment_interfaces_mocks.MockPaymentSpendPolicyProviderInterface
}

func WhenGettingPaymentSpendPolicyBeforeEach(t *testing.T) *WhenGettingPaymentSpendPolicyTestingSuite {
	mockController := gomock.NewController(t)

	mockPaymentSpendPolicyProvider := payment_interfaces_mocks.NewMockPaymentSpendPolicyProviderInterface(
		mockController,
	)
	mockPaymentSpendRuleRepository := payment_repositories_mocks.NewMockPaymentSpendRuleRepositoryInterface(
		mockController,
	)
	strategies, _ := newMockSpendStrategies(mockController)

	return &WhenGettingPaymentSpendPolicyTestingSuite{
		sut: use_cases.NewGetPaymentSpendPolicy(
			mockPaymentSpendPolicyProvider,
			mockPaymentSpendRuleRepository,
			strategies,
		),

		mockPaymentSpendPolicyProvider: mockPaymentSpendPolicyProvider,
	}
}

func TestWhenGettingPaymentSpendPolicy(t *testing.T) {
	t.Parallel()

	t.Run("Given the policy is materialized", func(t *testing.T) {
		t.Parallel()

		initSuite := func(suite *WhenGettingPaymentSpendPolicyTestingSuite) {
			policy := model_builders.NewPaymentSpendPolicyBuilder().Build()

			suite.mockPaymentSpendPolicyProvider.EXPECT().Ensure(gomock.Any()).Return(policy, nil).Times(1)
		}

		t.Run("Should return the policy head", func(t *testing.T) {
			t.Parallel()

			suite := WhenGettingPaymentSpendPolicyBeforeEach(t)
			initSuite(suite)

			response, err := suite.sut.Execute(context.Background())

			assert.NoError(t, err)
			assert.NotNil(t, response)
		})
	})

	t.Run("Given materializing the policy fails", func(t *testing.T) {
		t.Parallel()

		initSuite := func(suite *WhenGettingPaymentSpendPolicyTestingSuite) {
			suite.mockPaymentSpendPolicyProvider.EXPECT().
				Ensure(gomock.Any()).
				Return(nil, apperrors.NewInternalError("ensure_failed", "")).
				Times(1)
		}

		t.Run("Should propagate the provider's error", func(t *testing.T) {
			t.Parallel()

			suite := WhenGettingPaymentSpendPolicyBeforeEach(t)
			initSuite(suite)

			response, err := suite.sut.Execute(context.Background())

			assert.Nil(t, response)
			var internalErr *apperrors.InternalError
			assert.ErrorAs(t, err, &internalErr)
		})
	})
}
