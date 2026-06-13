package use_cases_test

import (
	"context"
	"testing"

	"github.com/6022-labs/agentic-agent-node/src/agent_payment/models"
	"github.com/6022-labs/agentic-agent-node/src/agent_payment/use_cases"
	apperrors "github.com/6022-labs/agentic-agent-node/src/common/errors"
	runtime_common_settings "github.com/6022-labs/agentic-agent-node/src/runtime_common/settings"
	identity_repositories_mocks "github.com/6022-labs/agentic-agent-node/tests/agent_identity_mocks/repositories_mocks"
	payment_repositories_mocks "github.com/6022-labs/agentic-agent-node/tests/agent_payment_mocks/repositories_mocks"
	"github.com/6022-labs/agentic-agent-node/tests/common_tests/model_builders"
	"github.com/ethereum/go-ethereum/common"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

type WhenGettingPaymentPolicyTestingSuite struct {
	sut *use_cases.GetPaymentPolicy

	mockAgentRepository         *identity_repositories_mocks.MockAgentRepositoryInterface
	mockPaymentPolicyRepository *payment_repositories_mocks.MockPaymentPolicyRepositoryInterface
	mockPaymentRuleRepository   *payment_repositories_mocks.MockPaymentRuleRepositoryInterface
}

func WhenGettingPaymentPolicyBeforeEach(t *testing.T) *WhenGettingPaymentPolicyTestingSuite {
	mockController := gomock.NewController(t)

	mockAgentRepository := identity_repositories_mocks.NewMockAgentRepositoryInterface(mockController)
	mockPaymentPolicyRepository := payment_repositories_mocks.NewMockPaymentPolicyRepositoryInterface(mockController)
	mockPaymentRuleRepository := payment_repositories_mocks.NewMockPaymentRuleRepositoryInterface(mockController)

	blockchainSettings := &runtime_common_settings.BlockchainSettings{Id: 80002}
	token6022Settings := &runtime_common_settings.Token6022Settings{
		SmartContractAddress: common.HexToAddress("0x4242424242424242424242424242424242424242"),
	}
	paymentFacilitatorsSettings := &runtime_common_settings.PaymentFacilitatorsSettings{
		"eip155:80002": "https://x402-amoy.polygon.technology",
	}

	return &WhenGettingPaymentPolicyTestingSuite{
		sut: use_cases.NewGetPaymentPolicy(mockAgentRepository, mockPaymentPolicyRepository, mockPaymentRuleRepository, blockchainSettings, token6022Settings, paymentFacilitatorsSettings),

		mockAgentRepository:         mockAgentRepository,
		mockPaymentPolicyRepository: mockPaymentPolicyRepository,
		mockPaymentRuleRepository:   mockPaymentRuleRepository,
	}
}

func TestWhenGettingPaymentPolicy(t *testing.T) {
	t.Parallel()

	t.Run("Given a policy exists", func(t *testing.T) {
		t.Parallel()

		initSuite := func(suite *WhenGettingPaymentPolicyTestingSuite) *models.PaymentPolicy {
			agent := model_builders.NewAgentBuilder().Build()
			policy := model_builders.NewPaymentPolicyBuilder().WithAgentId(agent.Id).Build()

			suite.mockAgentRepository.EXPECT().Get(gomock.Any()).Return(agent, nil).Times(1)
			suite.mockPaymentPolicyRepository.EXPECT().Get(gomock.Any()).Return(policy, nil).Times(1)
			return policy
		}

		t.Run("Should return the head without touching the rules", func(t *testing.T) {
			t.Parallel()

			suite := WhenGettingPaymentPolicyBeforeEach(t)
			policy := initSuite(suite)

			result, err := suite.sut.Execute(context.Background())

			assert.NoError(t, err)
			assert.Equal(t, policy.Id, result.Id)
			assert.NotEmpty(t, result.AgentWalletAddress)
		})
	})

	t.Run("Given no policy exists yet", func(t *testing.T) {
		t.Parallel()

		initSuite := func(suite *WhenGettingPaymentPolicyTestingSuite) {
			agent := model_builders.NewAgentBuilder().Build()

			suite.mockAgentRepository.EXPECT().Get(gomock.Any()).Return(agent, nil).Times(1)
			suite.mockPaymentPolicyRepository.EXPECT().Get(gomock.Any()).Return(nil, nil).Times(1)
			suite.mockPaymentPolicyRepository.EXPECT().
				Create(gomock.Any(), gomock.Any()).
				DoAndReturn(func(_ context.Context, policy *models.PaymentPolicy) (*models.PaymentPolicy, error) {
					assert.Equal(t, agent.Id, policy.AgentId)
					return policy, nil
				}).
				Times(1)
			suite.mockPaymentRuleRepository.EXPECT().
				Create(gomock.Any(), gomock.Any()).
				DoAndReturn(func(_ context.Context, rule *models.PaymentRule) (*models.PaymentRule, error) {
					// The seeded default: free access for the agent wallet (+ owner),
					// on the node's chain with the $6022 token as signing context.
					assert.Equal(t, "0", rule.Price)
					assert.Equal(t, "Owner free access", rule.Name)
					assert.Equal(t, "eip155:80002", rule.Network)
					assert.Equal(t, "0x4242424242424242424242424242424242424242", rule.Asset)
					return rule, nil
				}).
				Times(1)
		}

		t.Run("Should materialize the policy with the seeded default rule", func(t *testing.T) {
			t.Parallel()

			suite := WhenGettingPaymentPolicyBeforeEach(t)
			initSuite(suite)

			result, err := suite.sut.Execute(context.Background())

			assert.NoError(t, err)
			assert.NotEmpty(t, result.AgentWalletAddress)
		})
	})

	t.Run("Given the agent identity is not initialized", func(t *testing.T) {
		t.Parallel()

		initSuite := func(suite *WhenGettingPaymentPolicyTestingSuite) {
			suite.mockAgentRepository.EXPECT().Get(gomock.Any()).Return(nil, nil).Times(1)
		}

		t.Run("Should return an internal error", func(t *testing.T) {
			t.Parallel()

			suite := WhenGettingPaymentPolicyBeforeEach(t)
			initSuite(suite)

			result, err := suite.sut.Execute(context.Background())

			assert.Nil(t, result)
			var internalErr *apperrors.InternalError
			assert.ErrorAs(t, err, &internalErr)
		})
	})
}
