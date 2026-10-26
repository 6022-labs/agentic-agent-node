package use_cases_test

import (
	"context"
	"testing"

	"github.com/6022-labs/agentic-agent-node/src/agent_payment/models"
	"github.com/6022-labs/agentic-agent-node/src/agent_payment/use_cases"
	apperrors "github.com/6022-labs/agentic-agent-node/src/common/errors"
	payment_repositories_mocks "github.com/6022-labs/agentic-agent-node/tests/agent_payment_mocks/repositories_mocks"
	payment_interfaces_mocks "github.com/6022-labs/agentic-agent-node/tests/agent_payment_mocks/services_mocks/interfaces_mocks"
	"github.com/6022-labs/agentic-agent-node/tests/common_tests/model_builders"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

type WhenGettingPaymentReceivePolicyTestingSuite struct {
	sut *use_cases.GetPaymentReceivePolicy

	mockPaymentReceivePolicyProvider *payment_interfaces_mocks.MockPaymentReceivePolicyProviderInterface
	mockPaymentReceiveRuleRepository *payment_repositories_mocks.MockPaymentReceiveRuleRepositoryInterface
	mockStrategy                     *payment_interfaces_mocks.MockPaymentReceiveMethodStrategyInterface
}

func WhenGettingPaymentReceivePolicyBeforeEach(t *testing.T) *WhenGettingPaymentReceivePolicyTestingSuite {
	mockController := gomock.NewController(t)

	mockPaymentReceivePolicyProvider := payment_interfaces_mocks.NewMockPaymentReceivePolicyProviderInterface(
		mockController,
	)
	mockPaymentReceiveRuleRepository := payment_repositories_mocks.NewMockPaymentReceiveRuleRepositoryInterface(
		mockController,
	)

	strategies, mockStrategy := newMockStrategies(mockController)

	return &WhenGettingPaymentReceivePolicyTestingSuite{
		sut: use_cases.NewGetPaymentReceivePolicy(
			mockPaymentReceivePolicyProvider,
			mockPaymentReceiveRuleRepository,
			strategies,
		),

		mockPaymentReceivePolicyProvider: mockPaymentReceivePolicyProvider,
		mockPaymentReceiveRuleRepository: mockPaymentReceiveRuleRepository,
		mockStrategy:                     mockStrategy,
	}
}

func TestWhenGettingPaymentReceivePolicy(t *testing.T) {
	t.Parallel()

	t.Run("Given the provider refuses because the agent is not minted", func(t *testing.T) {
		t.Parallel()

		initSuite := func(suite *WhenGettingPaymentReceivePolicyTestingSuite) {
			suite.mockPaymentReceivePolicyProvider.EXPECT().
				Ensure(gomock.Any()).
				Return(nil, nil, apperrors.NewConflictError("agent_not_minted", "")).
				Times(1)
		}

		t.Run("Should propagate the conflict", func(t *testing.T) {
			t.Parallel()

			suite := WhenGettingPaymentReceivePolicyBeforeEach(t)
			initSuite(suite)

			result, err := suite.sut.Execute(context.Background())

			assert.Nil(t, result)
			var conflictErr *apperrors.ConflictError
			assert.ErrorAs(t, err, &conflictErr)
			assert.Equal(t, "agent_not_minted", conflictErr.Code)
		})
	})

	t.Run("Given reading the agent's wallet fails", func(t *testing.T) {
		t.Parallel()

		initSuite := func(suite *WhenGettingPaymentReceivePolicyTestingSuite) {
			agent := model_builders.NewAgentBuilder().Build()
			policy := model_builders.NewPaymentReceivePolicyBuilder().WithAgentId(agent.Id).Build()

			suite.mockPaymentReceivePolicyProvider.EXPECT().
				Ensure(gomock.Any()).
				Return(policy, agent, nil).
				Times(1)
			suite.mockPaymentReceivePolicyProvider.EXPECT().
				EvmAddress(gomock.Any(), agent.Id).
				Return("", assert.AnError).
				Times(1)
		}

		t.Run("Should return an internal error", func(t *testing.T) {
			t.Parallel()

			suite := WhenGettingPaymentReceivePolicyBeforeEach(t)
			initSuite(suite)

			result, err := suite.sut.Execute(context.Background())

			assert.Nil(t, result)
			var internalErr *apperrors.InternalError
			assert.ErrorAs(t, err, &internalErr)
		})
	})

	t.Run("Given the policy is materialized", func(t *testing.T) {
		t.Parallel()

		initSuite := func(suite *WhenGettingPaymentReceivePolicyTestingSuite) *models.PaymentReceivePolicy {
			agent := model_builders.NewAgentBuilder().Build()
			policy := model_builders.NewPaymentReceivePolicyBuilder().WithAgentId(agent.Id).Build()

			suite.mockPaymentReceivePolicyProvider.EXPECT().
				Ensure(gomock.Any()).
				Return(policy, agent, nil).
				Times(1)
			suite.mockPaymentReceivePolicyProvider.EXPECT().
				EvmAddress(gomock.Any(), agent.Id).
				Return("0xf39Fd6e51aad88F6F4ce6aB8827279cffFb92266", nil).
				Times(1)
			return policy
		}

		t.Run("Should return the head with the agent wallet to pay", func(t *testing.T) {
			t.Parallel()

			suite := WhenGettingPaymentReceivePolicyBeforeEach(t)
			policy := initSuite(suite)

			result, err := suite.sut.Execute(context.Background())

			assert.NoError(t, err)
			assert.Equal(t, policy.Id, result.Id)
			assert.NotEmpty(t, result.AgentWalletAddress)
		})
	})
}
