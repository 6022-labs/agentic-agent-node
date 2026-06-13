package use_cases_test

import (
	"context"
	"encoding/json"
	"errors"
	"testing"

	"github.com/6022-labs/agentic-agent-node/src/agent_payment/models"
	"github.com/6022-labs/agentic-agent-node/src/agent_payment/use_cases"
	apperrors "github.com/6022-labs/agentic-agent-node/src/common/errors"
	runtime_common_settings "github.com/6022-labs/agentic-agent-node/src/runtime_common/settings"
	identity_repositories_mocks "github.com/6022-labs/agentic-agent-node/tests/agent_identity_mocks/repositories_mocks"
	payment_repositories_mocks "github.com/6022-labs/agentic-agent-node/tests/agent_payment_mocks/repositories_mocks"
	payment_interfaces_mocks "github.com/6022-labs/agentic-agent-node/tests/agent_payment_mocks/services_mocks/interfaces_mocks"
	"github.com/6022-labs/agentic-agent-node/tests/common_tests/model_builders"
	"github.com/ethereum/go-ethereum/common"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

type WhenListingPaymentRulesTestingSuite struct {
	sut *use_cases.ListPaymentRules

	mockAgentRepository         *identity_repositories_mocks.MockAgentRepositoryInterface
	mockPaymentPolicyRepository *payment_repositories_mocks.MockPaymentPolicyRepositoryInterface
	mockPaymentRuleRepository   *payment_repositories_mocks.MockPaymentRuleRepositoryInterface
	mockStrategy                *payment_interfaces_mocks.MockPaymentMethodStrategyInterface
}

func WhenListingPaymentRulesBeforeEach(t *testing.T) *WhenListingPaymentRulesTestingSuite {
	mockController := gomock.NewController(t)

	mockAgentRepository := identity_repositories_mocks.NewMockAgentRepositoryInterface(mockController)
	mockPaymentPolicyRepository := payment_repositories_mocks.NewMockPaymentPolicyRepositoryInterface(mockController)
	mockPaymentRuleRepository := payment_repositories_mocks.NewMockPaymentRuleRepositoryInterface(mockController)

	blockchainSettings := &runtime_common_settings.BlockchainSettings{Id: 80002}
	token6022Settings := &runtime_common_settings.Token6022Settings{
		SmartContractAddress: common.HexToAddress("0x4242424242424242424242424242424242424242"),
	}
	strategies, mockStrategy := newMockStrategies(mockController)

	return &WhenListingPaymentRulesTestingSuite{
		sut: use_cases.NewListPaymentRules(mockAgentRepository, mockPaymentPolicyRepository, mockPaymentRuleRepository, strategies, blockchainSettings, token6022Settings),

		mockAgentRepository:         mockAgentRepository,
		mockPaymentPolicyRepository: mockPaymentPolicyRepository,
		mockPaymentRuleRepository:   mockPaymentRuleRepository,
		mockStrategy:                mockStrategy,
	}
}

func TestWhenListingPaymentRules(t *testing.T) {
	t.Parallel()

	t.Run("Given a policy with rules exists", func(t *testing.T) {
		t.Parallel()

		initSuite := func(suite *WhenListingPaymentRulesTestingSuite) *models.PaymentRule {
			agent := model_builders.NewAgentBuilder().Build()
			policy := model_builders.NewPaymentPolicyBuilder().WithAgentId(agent.Id).Build()
			rule := model_builders.NewPaymentRuleBuilder().WithPolicyId(policy.Id).Build()

			suite.mockAgentRepository.EXPECT().Get(gomock.Any()).Return(agent, nil).Times(1)
			suite.mockPaymentPolicyRepository.EXPECT().Get(gomock.Any()).Return(policy, nil).Times(1)
			suite.mockPaymentRuleRepository.EXPECT().List(gomock.Any()).Return([]models.PaymentRule{*rule}, nil).Times(1)
			// Each rule's config is joined back from its method's table.
			suite.mockStrategy.EXPECT().
				LoadConfig(gomock.Any(), rule.Id).
				Return(json.RawMessage(`{"network":"eip155:80002"}`), nil).
				Times(1)
			return rule
		}

		t.Run("Should return the rules with their config", func(t *testing.T) {
			t.Parallel()

			suite := WhenListingPaymentRulesBeforeEach(t)
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

	t.Run("Given no policy exists yet", func(t *testing.T) {
		t.Parallel()

		initSuite := func(suite *WhenListingPaymentRulesTestingSuite) {
			agent := model_builders.NewAgentBuilder().Build()
			policy := model_builders.NewPaymentPolicyBuilder().WithAgentId(agent.Id).Build()

			suite.mockAgentRepository.EXPECT().Get(gomock.Any()).Return(agent, nil).Times(1)
			suite.mockPaymentPolicyRepository.EXPECT().Get(gomock.Any()).Return(nil, nil).Times(1)
			suite.mockPaymentPolicyRepository.EXPECT().Create(gomock.Any(), gomock.Any()).Return(policy, nil).Times(1)
			// Seeding the default rule persists its config, then the empty list
			// is returned.
			suite.mockPaymentRuleRepository.EXPECT().Create(gomock.Any(), gomock.Any()).Return(model_builders.NewPaymentRuleBuilder().Build(), nil).Times(1)
			suite.mockStrategy.EXPECT().SaveConfig(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil).Times(1)
			suite.mockPaymentRuleRepository.EXPECT().List(gomock.Any()).Return([]models.PaymentRule{}, nil).Times(1)
		}

		t.Run("Should materialize the policy before listing", func(t *testing.T) {
			t.Parallel()

			suite := WhenListingPaymentRulesBeforeEach(t)
			initSuite(suite)

			result, err := suite.sut.Execute(context.Background())

			assert.NoError(t, err)
			assert.NotNil(t, result)
		})
	})

	t.Run("Given listing the rules fails", func(t *testing.T) {
		t.Parallel()

		initSuite := func(suite *WhenListingPaymentRulesTestingSuite) {
			agent := model_builders.NewAgentBuilder().Build()
			policy := model_builders.NewPaymentPolicyBuilder().WithAgentId(agent.Id).Build()

			suite.mockAgentRepository.EXPECT().Get(gomock.Any()).Return(agent, nil).Times(1)
			suite.mockPaymentPolicyRepository.EXPECT().Get(gomock.Any()).Return(policy, nil).Times(1)
			suite.mockPaymentRuleRepository.EXPECT().List(gomock.Any()).Return(nil, errors.New("db down")).Times(1)
		}

		t.Run("Should return an internal error", func(t *testing.T) {
			t.Parallel()

			suite := WhenListingPaymentRulesBeforeEach(t)
			initSuite(suite)

			result, err := suite.sut.Execute(context.Background())

			assert.Nil(t, result)
			var internalErr *apperrors.InternalError
			assert.ErrorAs(t, err, &internalErr)
		})
	})
}
