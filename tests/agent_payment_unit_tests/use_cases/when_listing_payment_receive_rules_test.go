package use_cases_test

import (
	"context"
	"encoding/json"
	"errors"
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

type WhenListingPaymentReceiveRulesTestingSuite struct {
	sut *use_cases.ListPaymentReceiveRules

	mockPaymentReceivePolicyProvider *payment_interfaces_mocks.MockPaymentReceivePolicyProviderInterface
	mockPaymentReceiveRuleRepository *payment_repositories_mocks.MockPaymentReceiveRuleRepositoryInterface
	mockStrategy                     *payment_interfaces_mocks.MockPaymentReceiveMethodStrategyInterface
}

func WhenListingPaymentReceiveRulesBeforeEach(t *testing.T) *WhenListingPaymentReceiveRulesTestingSuite {
	mockController := gomock.NewController(t)

	mockPaymentReceivePolicyProvider := payment_interfaces_mocks.NewMockPaymentReceivePolicyProviderInterface(
		mockController,
	)
	mockPaymentReceiveRuleRepository := payment_repositories_mocks.NewMockPaymentReceiveRuleRepositoryInterface(
		mockController,
	)

	strategies, mockStrategy := newMockStrategies(mockController)

	return &WhenListingPaymentReceiveRulesTestingSuite{
		sut: use_cases.NewListPaymentReceiveRules(
			mockPaymentReceivePolicyProvider,
			mockPaymentReceiveRuleRepository,
			strategies,
		),

		mockPaymentReceivePolicyProvider: mockPaymentReceivePolicyProvider,
		mockPaymentReceiveRuleRepository: mockPaymentReceiveRuleRepository,
		mockStrategy:                     mockStrategy,
	}
}

func TestWhenListingPaymentReceiveRules(t *testing.T) {
	t.Parallel()

	t.Run("Given the provider refuses because the agent is not minted", func(t *testing.T) {
		t.Parallel()

		initSuite := func(suite *WhenListingPaymentReceiveRulesTestingSuite) {
			suite.mockPaymentReceivePolicyProvider.EXPECT().
				Ensure(gomock.Any()).
				Return(nil, nil, apperrors.NewConflictError("agent_not_minted", "")).
				Times(1)
		}

		t.Run("Should propagate the conflict without listing rules", func(t *testing.T) {
			t.Parallel()

			suite := WhenListingPaymentReceiveRulesBeforeEach(t)
			initSuite(suite)

			result, err := suite.sut.Execute(context.Background())

			assert.Nil(t, result)
			var conflictErr *apperrors.ConflictError
			assert.ErrorAs(t, err, &conflictErr)
		})
	})

	t.Run("Given listing the rules fails", func(t *testing.T) {
		t.Parallel()

		initSuite := func(suite *WhenListingPaymentReceiveRulesTestingSuite) {
			agent := model_builders.NewAgentBuilder().Build()
			policy := model_builders.NewPaymentReceivePolicyBuilder().WithAgentId(agent.Id).Build()

			suite.mockPaymentReceivePolicyProvider.EXPECT().Ensure(gomock.Any()).Return(policy, agent, nil).Times(1)
			suite.mockPaymentReceiveRuleRepository.EXPECT().
				ListByPolicyId(gomock.Any(), gomock.Any()).
				Return(nil, errors.New("db down")).
				Times(1)
		}

		t.Run("Should return an internal error", func(t *testing.T) {
			t.Parallel()

			suite := WhenListingPaymentReceiveRulesBeforeEach(t)
			initSuite(suite)

			result, err := suite.sut.Execute(context.Background())

			assert.Nil(t, result)
			var internalErr *apperrors.InternalError
			assert.ErrorAs(t, err, &internalErr)
		})
	})

	t.Run("Given a policy with rules exists", func(t *testing.T) {
		t.Parallel()

		initSuite := func(suite *WhenListingPaymentReceiveRulesTestingSuite) *models.PaymentReceiveRule {
			agent := model_builders.NewAgentBuilder().Build()
			policy := model_builders.NewPaymentReceivePolicyBuilder().WithAgentId(agent.Id).Build()
			rule := model_builders.NewPaymentReceiveRuleBuilder().WithPolicyId(policy.Id).Build()

			suite.mockPaymentReceivePolicyProvider.EXPECT().Ensure(gomock.Any()).Return(policy, agent, nil).Times(1)
			suite.mockPaymentReceiveRuleRepository.EXPECT().
				ListByPolicyId(gomock.Any(), gomock.Any()).
				Return([]models.PaymentReceiveRule{*rule}, nil).
				Times(1)
			// Each rule's config is joined back from its method's table.
			suite.mockStrategy.EXPECT().
				LoadConfig(gomock.Any(), rule.Id).
				Return(json.RawMessage(`{"network":"eip155:80002"}`), nil).
				Times(1)
			return rule
		}

		t.Run("Should return the rules with their config", func(t *testing.T) {
			t.Parallel()

			suite := WhenListingPaymentReceiveRulesBeforeEach(t)
			rule := initSuite(suite)

			result, err := suite.sut.Execute(context.Background())

			assert.NoError(t, err)
			if assert.Len(t, result.Rules, 1) {
				assert.Equal(t, rule.Id, result.Rules[0].Id)
				assert.Equal(t, rule.Name, result.Rules[0].Name)
				assert.JSONEq(t, `{"network":"eip155:80002"}`, string(result.Rules[0].Config))
			}
		})
	})
}
