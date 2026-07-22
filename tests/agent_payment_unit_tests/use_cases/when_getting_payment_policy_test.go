package use_cases_test

import (
	"context"
	"encoding/json"
	"testing"

	identity_models "github.com/6022-labs/agentic-agent-node/src/agent_identity/models"
	identity_enums "github.com/6022-labs/agentic-agent-node/src/agent_identity/models/enums"
	"github.com/6022-labs/agentic-agent-node/src/agent_payment/models"
	"github.com/6022-labs/agentic-agent-node/src/agent_payment/models/enums"
	"github.com/6022-labs/agentic-agent-node/src/agent_payment/use_cases"
	apperrors "github.com/6022-labs/agentic-agent-node/src/common/errors"
	runtime_common_settings "github.com/6022-labs/agentic-agent-node/src/runtime_common/settings"
	identity_repositories_mocks "github.com/6022-labs/agentic-agent-node/tests/agent_identity_mocks/repositories_mocks"
	payment_repositories_mocks "github.com/6022-labs/agentic-agent-node/tests/agent_payment_mocks/repositories_mocks"
	payment_interfaces_mocks "github.com/6022-labs/agentic-agent-node/tests/agent_payment_mocks/services_mocks/interfaces_mocks"
	"github.com/6022-labs/agentic-agent-node/tests/common_tests/model_builders"
	"github.com/ethereum/go-ethereum/common"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

type WhenGettingPaymentPolicyTestingSuite struct {
	sut *use_cases.GetPaymentPolicy

	mockAgentRepository         *identity_repositories_mocks.MockAgentRepositoryInterface
	mockAgentAddressRepository  *identity_repositories_mocks.MockAgentAddressRepositoryInterface
	mockPaymentPolicyRepository *payment_repositories_mocks.MockPaymentPolicyRepositoryInterface
	mockPaymentRuleRepository   *payment_repositories_mocks.MockPaymentRuleRepositoryInterface
	mockStrategy                *payment_interfaces_mocks.MockPaymentMethodStrategyInterface
}

func WhenGettingPaymentPolicyBeforeEach(t *testing.T) *WhenGettingPaymentPolicyTestingSuite {
	mockController := gomock.NewController(t)

	mockAgentRepository := identity_repositories_mocks.NewMockAgentRepositoryInterface(mockController)
	mockAgentAddressRepository := identity_repositories_mocks.NewMockAgentAddressRepositoryInterface(mockController)
	mockPaymentPolicyRepository := payment_repositories_mocks.NewMockPaymentPolicyRepositoryInterface(mockController)
	mockPaymentRuleRepository := payment_repositories_mocks.NewMockPaymentRuleRepositoryInterface(mockController)

	token6022Settings := &runtime_common_settings.Token6022Settings{
		"80002": common.HexToAddress("0x4242424242424242424242424242424242424242"),
	}
	strategies, mockStrategy := newMockStrategies(mockController)

	return &WhenGettingPaymentPolicyTestingSuite{
		sut: use_cases.NewGetPaymentPolicy(mockAgentRepository, mockAgentAddressRepository, mockPaymentPolicyRepository, mockPaymentRuleRepository, strategies, token6022Settings),

		mockAgentRepository:         mockAgentRepository,
		mockAgentAddressRepository:  mockAgentAddressRepository,
		mockPaymentPolicyRepository: mockPaymentPolicyRepository,
		mockPaymentRuleRepository:   mockPaymentRuleRepository,
		mockStrategy:                mockStrategy,
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
			suite.mockAgentAddressRepository.EXPECT().GetByAgentIdAndType(gomock.Any(), gomock.Any(), identity_enums.AddressTypeEvm).Return(&identity_models.AgentAddress{Address: "0x2f0DDd93040C19863527b0553fECeB696862EA7D"}, nil).AnyTimes()
			suite.mockPaymentPolicyRepository.EXPECT().GetByAgentId(gomock.Any(), gomock.Any()).Return(policy, nil).Times(1)
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
			suite.mockAgentAddressRepository.EXPECT().GetByAgentIdAndType(gomock.Any(), gomock.Any(), identity_enums.AddressTypeEvm).Return(&identity_models.AgentAddress{Address: "0x2f0DDd93040C19863527b0553fECeB696862EA7D"}, nil).AnyTimes()
			suite.mockPaymentPolicyRepository.EXPECT().GetByAgentId(gomock.Any(), gomock.Any()).Return(nil, nil).Times(1)
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
					// The seeded default: free access for the agent wallet (+ owner).
					assert.Equal(t, "0", rule.Price)
					assert.Equal(t, "Owner free access", rule.Name)
					assert.Equal(t, enums.PaymentMethodX402, rule.Method)
					return rule, nil
				}).
				Times(1)
			// The default rule's x402 config is persisted on the node's chain
			// with the $6022 token as the signing context.
			suite.mockStrategy.EXPECT().
				SaveConfig(gomock.Any(), gomock.Any(), gomock.Any()).
				DoAndReturn(func(_ context.Context, _ uuid.UUID, raw json.RawMessage) error {
					config := &models.PaymentRuleX402Config{}
					assert.NoError(t, config.Parse(raw))
					assert.Equal(t, "eip155:80002", config.Network)
					assert.Equal(t, "0x4242424242424242424242424242424242424242", config.Asset)
					return nil
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

	t.Run("Given the policy creation races with a concurrent request", func(t *testing.T) {
		t.Parallel()

		initSuite := func(suite *WhenGettingPaymentPolicyTestingSuite) *models.PaymentPolicy {
			agent := model_builders.NewAgentBuilder().Build()
			policy := model_builders.NewPaymentPolicyBuilder().WithAgentId(agent.Id).Build()

			suite.mockAgentRepository.EXPECT().Get(gomock.Any()).Return(agent, nil).Times(1)
			suite.mockAgentAddressRepository.EXPECT().GetByAgentIdAndType(gomock.Any(), gomock.Any(), identity_enums.AddressTypeEvm).Return(&identity_models.AgentAddress{Address: "0x2f0DDd93040C19863527b0553fECeB696862EA7D"}, nil).AnyTimes()
			gomock.InOrder(
				suite.mockPaymentPolicyRepository.EXPECT().GetByAgentId(gomock.Any(), gomock.Any()).Return(nil, nil).Times(1),
				suite.mockPaymentPolicyRepository.EXPECT().Create(gomock.Any(), gomock.Any()).Return(nil, assert.AnError).Times(1),
				suite.mockPaymentPolicyRepository.EXPECT().GetByAgentId(gomock.Any(), gomock.Any()).Return(policy, nil).Times(1),
			)
			return policy
		}

		t.Run("Should return the concurrently created policy without seeding rules", func(t *testing.T) {
			t.Parallel()

			suite := WhenGettingPaymentPolicyBeforeEach(t)
			policy := initSuite(suite)

			result, err := suite.sut.Execute(context.Background())

			assert.NoError(t, err)
			assert.Equal(t, policy.Id, result.Id)
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
