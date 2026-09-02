package use_cases_test

import (
	"context"
	"encoding/json"
	"errors"
	"testing"

	identity_enums "github.com/6022-labs/agentic-agent-node/src/agent_identity/models/enums"
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

type WhenListingPaymentReceiveRulesTestingSuite struct {
	sut *use_cases.ListPaymentReceiveRules

	mockAgentRepository                *identity_repositories_mocks.MockAgentRepositoryInterface
	mockAgentAddressRepository         *identity_repositories_mocks.MockAgentAddressRepositoryInterface
	mockPaymentReceivePolicyRepository *payment_repositories_mocks.MockPaymentReceivePolicyRepositoryInterface
	mockPaymentReceiveRuleRepository   *payment_repositories_mocks.MockPaymentReceiveRuleRepositoryInterface
	mockStrategy                       *payment_interfaces_mocks.MockPaymentReceiveMethodStrategyInterface
}

func WhenListingPaymentReceiveRulesBeforeEach(t *testing.T) *WhenListingPaymentReceiveRulesTestingSuite {
	mockController := gomock.NewController(t)

	mockAgentRepository := identity_repositories_mocks.NewMockAgentRepositoryInterface(mockController)
	mockAgentAddressRepository := identity_repositories_mocks.NewMockAgentAddressRepositoryInterface(mockController)
	mockPaymentReceivePolicyRepository := payment_repositories_mocks.NewMockPaymentReceivePolicyRepositoryInterface(
		mockController,
	)
	mockPaymentReceiveRuleRepository := payment_repositories_mocks.NewMockPaymentReceiveRuleRepositoryInterface(
		mockController,
	)

	token6022Settings := &runtime_common_settings.Token6022Settings{
		"80002": common.HexToAddress("0x4242424242424242424242424242424242424242"),
	}
	strategies, mockStrategy := newMockStrategies(mockController)

	return &WhenListingPaymentReceiveRulesTestingSuite{
		sut: use_cases.NewListPaymentReceiveRules(
			mockAgentRepository,
			mockAgentAddressRepository,
			mockPaymentReceivePolicyRepository,
			mockPaymentReceiveRuleRepository,
			strategies,
			token6022Settings,
		),

		mockAgentRepository:                mockAgentRepository,
		mockAgentAddressRepository:         mockAgentAddressRepository,
		mockPaymentReceivePolicyRepository: mockPaymentReceivePolicyRepository,
		mockPaymentReceiveRuleRepository:   mockPaymentReceiveRuleRepository,
		mockStrategy:                       mockStrategy,
	}
}

func TestWhenListingPaymentReceiveRules(t *testing.T) {
	t.Parallel()

	t.Run("Given the agent is not minted yet", func(t *testing.T) {
		t.Parallel()

		initSuite := func(suite *WhenListingPaymentReceiveRulesTestingSuite) {
			agent := model_builders.NewAgentBuilder().WithoutCollectionTokenId().Build()
			suite.mockAgentRepository.EXPECT().Get(gomock.Any()).Return(agent, nil).Times(1)
		}

		t.Run("Should return a conflict without touching the policy", func(t *testing.T) {
			t.Parallel()

			suite := WhenListingPaymentReceiveRulesBeforeEach(t)
			initSuite(suite)

			result, err := suite.sut.Execute(context.Background())

			assert.Nil(t, result)
			var conflictErr *apperrors.ConflictError
			assert.ErrorAs(t, err, &conflictErr)
			assert.Equal(t, "agent_not_minted", conflictErr.Code)
		})
	})

	t.Run("Given listing the rules fails", func(t *testing.T) {
		t.Parallel()

		initSuite := func(suite *WhenListingPaymentReceiveRulesTestingSuite) {
			agent := model_builders.NewAgentBuilder().Build()
			policy := model_builders.NewPaymentReceivePolicyBuilder().WithAgentId(agent.Id).Build()

			suite.mockAgentRepository.EXPECT().Get(gomock.Any()).Return(agent, nil).Times(1)
			suite.mockAgentAddressRepository.EXPECT().
				GetByAgentIdAndType(gomock.Any(), gomock.Any(), identity_enums.AddressTypeEvm).
				Return(nil, nil).
				AnyTimes()
			suite.mockPaymentReceivePolicyRepository.EXPECT().
				GetByAgentId(gomock.Any(), gomock.Any()).
				Return(policy, nil).
				Times(1)
			suite.mockPaymentReceiveRuleRepository.EXPECT().
				ListByPolicyId(gomock.Any(), gomock.Any()).
				Return(nil, errors.New("db down")).
				Times(1)
		}

		t.Run("Should return an InternalError", func(t *testing.T) {
			t.Parallel()

			suite := WhenListingPaymentReceiveRulesBeforeEach(t)
			initSuite(suite)

			result, err := suite.sut.Execute(context.Background())

			assert.Nil(t, result)
			var internalErr *apperrors.InternalError
			assert.ErrorAs(t, err, &internalErr)
		})
	})

	t.Run("Given no policy exists", func(t *testing.T) {
		t.Parallel()

		initSuite := func(suite *WhenListingPaymentReceiveRulesTestingSuite) {
			agent := model_builders.NewAgentBuilder().Build()
			policy := model_builders.NewPaymentReceivePolicyBuilder().WithAgentId(agent.Id).Build()

			suite.mockAgentRepository.EXPECT().Get(gomock.Any()).Return(agent, nil).Times(1)
			suite.mockAgentAddressRepository.EXPECT().
				GetByAgentIdAndType(gomock.Any(), gomock.Any(), identity_enums.AddressTypeEvm).
				Return(nil, nil).
				AnyTimes()
			suite.mockPaymentReceivePolicyRepository.EXPECT().
				GetByAgentId(gomock.Any(), gomock.Any()).
				Return(nil, nil).
				Times(1)
			suite.mockPaymentReceivePolicyRepository.EXPECT().
				Create(gomock.Any(), gomock.Any()).
				Return(policy, nil).
				Times(1)
			// Seeding the default rule persists its config, then the empty list is returned.
			suite.mockPaymentReceiveRuleRepository.EXPECT().
				Create(gomock.Any(), gomock.Any()).
				Return(model_builders.NewPaymentReceiveRuleBuilder().Build(), nil).
				Times(1)
			suite.mockStrategy.EXPECT().SaveConfig(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil).Times(1)
			suite.mockPaymentReceiveRuleRepository.EXPECT().
				ListByPolicyId(gomock.Any(), gomock.Any()).
				Return([]models.PaymentReceiveRule{}, nil).
				Times(1)
		}

		t.Run("Should materialize the policy before listing", func(t *testing.T) {
			t.Parallel()

			suite := WhenListingPaymentReceiveRulesBeforeEach(t)
			initSuite(suite)

			result, err := suite.sut.Execute(context.Background())

			assert.NoError(t, err)
			assert.NotNil(t, result)
		})
	})

	t.Run("Given a policy with rules exists", func(t *testing.T) {
		t.Parallel()

		initSuite := func(suite *WhenListingPaymentReceiveRulesTestingSuite) *models.PaymentReceiveRule {
			agent := model_builders.NewAgentBuilder().Build()
			policy := model_builders.NewPaymentReceivePolicyBuilder().WithAgentId(agent.Id).Build()
			rule := model_builders.NewPaymentReceiveRuleBuilder().WithPolicyId(policy.Id).Build()

			suite.mockAgentRepository.EXPECT().Get(gomock.Any()).Return(agent, nil).Times(1)
			suite.mockAgentAddressRepository.EXPECT().
				GetByAgentIdAndType(gomock.Any(), gomock.Any(), identity_enums.AddressTypeEvm).
				Return(nil, nil).
				AnyTimes()
			suite.mockPaymentReceivePolicyRepository.EXPECT().
				GetByAgentId(gomock.Any(), gomock.Any()).
				Return(policy, nil).
				Times(1)
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
