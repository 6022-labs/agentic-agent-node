package use_cases_test

import (
	"context"
	"testing"

	"github.com/6022-labs/agentic-agent-node/src/agent_payment/models"
	"github.com/6022-labs/agentic-agent-node/src/agent_payment/use_cases"
	"github.com/6022-labs/agentic-agent-node/src/agent_payment/use_cases/requests"
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

type WhenCreatingPaymentRuleTestingSuite struct {
	sut *use_cases.CreatePaymentRule

	mockAgentRepository         *identity_repositories_mocks.MockAgentRepositoryInterface
	mockPaymentPolicyRepository *payment_repositories_mocks.MockPaymentPolicyRepositoryInterface
	mockPaymentRuleRepository   *payment_repositories_mocks.MockPaymentRuleRepositoryInterface
	mockErc20Validator          *payment_interfaces_mocks.MockErc20ValidatorInterface
}

func WhenCreatingPaymentRuleBeforeEach(t *testing.T) *WhenCreatingPaymentRuleTestingSuite {
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
	mockErc20Validator := payment_interfaces_mocks.NewMockErc20ValidatorInterface(mockController)

	return &WhenCreatingPaymentRuleTestingSuite{
		sut: use_cases.NewCreatePaymentRule(mockAgentRepository, mockPaymentPolicyRepository, mockPaymentRuleRepository, blockchainSettings, token6022Settings, paymentFacilitatorsSettings, mockErc20Validator),

		mockAgentRepository:         mockAgentRepository,
		mockPaymentPolicyRepository: mockPaymentPolicyRepository,
		mockPaymentRuleRepository:   mockPaymentRuleRepository,
		mockErc20Validator:          mockErc20Validator,
	}
}

func generateSavePaymentRuleRequest() *requests.SavePaymentRuleRequest {
	return &requests.SavePaymentRuleRequest{
		Name:              "Public price",
		Network:           "eip155:80002",
		Asset:             "0x41E94Eb019C0762f9Bfcf9Fb1E58725BfB0e7582",
		PayTo:             "0xf39Fd6e51aad88F6F4ce6aB8827279cffFb92266",
		Price:             "10000",
		MaxTimeoutSeconds: 60,
		Eip712Name:        "USDC",
		Eip712Version:     "2",
		AddressPatterns:   []string{".*"},
	}
}

func TestWhenCreatingPaymentRule(t *testing.T) {
	t.Parallel()

	t.Run("Given the asset is a valid ERC-20", func(t *testing.T) {
		t.Parallel()

		initSuite := func(suite *WhenCreatingPaymentRuleTestingSuite) *models.PaymentPolicy {
			agent := model_builders.NewAgentBuilder().Build()
			policy := model_builders.NewPaymentPolicyBuilder().WithAgentId(agent.Id).Build()

			suite.mockAgentRepository.EXPECT().Get(gomock.Any()).Return(agent, nil).Times(1)
			suite.mockPaymentPolicyRepository.EXPECT().Get(gomock.Any()).Return(policy, nil).Times(1)
			suite.mockErc20Validator.EXPECT().
				ValidateErc20(gomock.Any(), int64(80002), "0x41E94Eb019C0762f9Bfcf9Fb1E58725BfB0e7582").
				Return(nil).
				Times(1)
			suite.mockPaymentRuleRepository.EXPECT().
				Create(gomock.Any(), gomock.Any()).
				DoAndReturn(func(_ context.Context, rule *models.PaymentRule) (*models.PaymentRule, error) {
					assert.Equal(t, policy.Id, rule.PolicyId)
					assert.Equal(t, "Public price", rule.Name)
					return rule, nil
				}).
				Times(1)
			return policy
		}

		t.Run("Should create the rule bound to the policy", func(t *testing.T) {
			t.Parallel()

			suite := WhenCreatingPaymentRuleBeforeEach(t)
			initSuite(suite)

			rule, err := suite.sut.Execute(context.Background(), generateSavePaymentRuleRequest())

			assert.NoError(t, err)
			assert.Equal(t, "Public price", rule.Name)
		})
	})

	t.Run("Given no policy exists yet", func(t *testing.T) {
		t.Parallel()

		initSuite := func(suite *WhenCreatingPaymentRuleTestingSuite) {
			agent := model_builders.NewAgentBuilder().Build()
			policy := model_builders.NewPaymentPolicyBuilder().WithAgentId(agent.Id).Build()

			suite.mockAgentRepository.EXPECT().Get(gomock.Any()).Return(agent, nil).Times(1)
			suite.mockPaymentPolicyRepository.EXPECT().Get(gomock.Any()).Return(nil, nil).Times(1)
			suite.mockPaymentPolicyRepository.EXPECT().Create(gomock.Any(), gomock.Any()).Return(policy, nil).Times(1)
			suite.mockErc20Validator.EXPECT().ValidateErc20(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil).Times(1)
			// First Create seeds the default rule; the second creates the request.
			suite.mockPaymentRuleRepository.EXPECT().
				Create(gomock.Any(), gomock.Any()).
				DoAndReturn(func(_ context.Context, rule *models.PaymentRule) (*models.PaymentRule, error) {
					return rule, nil
				}).
				Times(2)
		}

		t.Run("Should materialize the policy before creating the rule", func(t *testing.T) {
			t.Parallel()

			suite := WhenCreatingPaymentRuleBeforeEach(t)
			initSuite(suite)

			rule, err := suite.sut.Execute(context.Background(), generateSavePaymentRuleRequest())

			assert.NoError(t, err)
			assert.Equal(t, "Public price", rule.Name)
		})
	})

	t.Run("Given a paid rule targets a network without a facilitator", func(t *testing.T) {
		t.Parallel()

		t.Run("Should return a validation error before touching the policy", func(t *testing.T) {
			t.Parallel()

			suite := WhenCreatingPaymentRuleBeforeEach(t)

			request := generateSavePaymentRuleRequest()
			request.Network = "eip155:8453"
			rule, err := suite.sut.Execute(context.Background(), request)

			assert.Nil(t, rule)
			var validationErr *apperrors.ValidationError
			assert.ErrorAs(t, err, &validationErr)
			assert.Equal(t, "network", validationErr.Field)
		})
	})

	t.Run("Given the request is invalid", func(t *testing.T) {
		t.Parallel()

		t.Run("Should return a validation error before touching the policy", func(t *testing.T) {
			t.Parallel()

			suite := WhenCreatingPaymentRuleBeforeEach(t)

			request := generateSavePaymentRuleRequest()
			request.Name = ""
			rule, err := suite.sut.Execute(context.Background(), request)

			assert.Nil(t, rule)
			var validationErr *apperrors.ValidationError
			assert.ErrorAs(t, err, &validationErr)
		})
	})

	t.Run("Given the asset is not an ERC-20 contract", func(t *testing.T) {
		t.Parallel()

		initSuite := func(suite *WhenCreatingPaymentRuleTestingSuite) {
			agent := model_builders.NewAgentBuilder().Build()
			policy := model_builders.NewPaymentPolicyBuilder().WithAgentId(agent.Id).Build()

			suite.mockAgentRepository.EXPECT().Get(gomock.Any()).Return(agent, nil).Times(1)
			suite.mockPaymentPolicyRepository.EXPECT().Get(gomock.Any()).Return(policy, nil).Times(1)
			suite.mockErc20Validator.EXPECT().
				ValidateErc20(gomock.Any(), gomock.Any(), gomock.Any()).
				Return(assert.AnError).
				Times(1)
		}

		t.Run("Should return a validation error on the asset field", func(t *testing.T) {
			t.Parallel()

			suite := WhenCreatingPaymentRuleBeforeEach(t)
			initSuite(suite)

			rule, err := suite.sut.Execute(context.Background(), generateSavePaymentRuleRequest())

			assert.Nil(t, rule)
			var validationErr *apperrors.ValidationError
			assert.ErrorAs(t, err, &validationErr)
			assert.Equal(t, "asset", validationErr.Field)
		})
	})
}
