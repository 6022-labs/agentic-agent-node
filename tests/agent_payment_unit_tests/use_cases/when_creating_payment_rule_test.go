package use_cases_test

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/6022-labs/agentic-agent-node/src/agent_payment/models"
	"github.com/6022-labs/agentic-agent-node/src/agent_payment/models/enums"
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
	mockStrategy                *payment_interfaces_mocks.MockPaymentMethodStrategyInterface
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
	strategies, mockStrategy := newMockStrategies(mockController)

	return &WhenCreatingPaymentRuleTestingSuite{
		sut: use_cases.NewCreatePaymentRule(mockAgentRepository, mockPaymentPolicyRepository, mockPaymentRuleRepository, strategies, blockchainSettings, token6022Settings),

		mockAgentRepository:         mockAgentRepository,
		mockPaymentPolicyRepository: mockPaymentPolicyRepository,
		mockPaymentRuleRepository:   mockPaymentRuleRepository,
		mockStrategy:                mockStrategy,
	}
}

func generateSavePaymentRuleRequest() *requests.SavePaymentRuleRequest {
	return &requests.SavePaymentRuleRequest{
		Name:   "Public price",
		Price:  "10000",
		Config: json.RawMessage(`{"network":"eip155:80002","asset":"0x41E94Eb019C0762f9Bfcf9Fb1E58725BfB0e7582","payTo":"0xf39Fd6e51aad88F6F4ce6aB8827279cffFb92266","addressPatterns":[".*"]}`),
	}
}

func TestWhenCreatingPaymentRule(t *testing.T) {
	t.Parallel()

	t.Run("Given the rule config is valid", func(t *testing.T) {
		t.Parallel()

		initSuite := func(suite *WhenCreatingPaymentRuleTestingSuite) *models.PaymentPolicy {
			agent := model_builders.NewAgentBuilder().Build()
			policy := model_builders.NewPaymentPolicyBuilder().WithAgentId(agent.Id).Build()

			suite.mockStrategy.EXPECT().Validate(gomock.Any(), gomock.Any()).Return(nil).Times(1)
			suite.mockAgentRepository.EXPECT().Get(gomock.Any()).Return(agent, nil).Times(1)
			suite.mockPaymentPolicyRepository.EXPECT().Get(gomock.Any()).Return(policy, nil).Times(1)
			suite.mockPaymentRuleRepository.EXPECT().
				Create(gomock.Any(), gomock.Any()).
				DoAndReturn(func(_ context.Context, rule *models.PaymentRule) (*models.PaymentRule, error) {
					assert.Equal(t, policy.Id, rule.PolicyId)
					assert.Equal(t, "Public price", rule.Name)
					assert.Equal(t, enums.PaymentMethodX402, rule.Method)
					return rule, nil
				}).
				Times(1)
			// The rule's config is persisted by the strategy, keyed by the rule.
			suite.mockStrategy.EXPECT().SaveConfig(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil).Times(1)
			return policy
		}

		t.Run("Should create the rule and persist its config", func(t *testing.T) {
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

			suite.mockStrategy.EXPECT().Validate(gomock.Any(), gomock.Any()).Return(nil).Times(1)
			suite.mockAgentRepository.EXPECT().Get(gomock.Any()).Return(agent, nil).Times(1)
			suite.mockPaymentPolicyRepository.EXPECT().Get(gomock.Any()).Return(nil, nil).Times(1)
			suite.mockPaymentPolicyRepository.EXPECT().Create(gomock.Any(), gomock.Any()).Return(policy, nil).Times(1)
			// One Create + SaveConfig seeds the default rule, the second pair
			// persists the request's rule.
			suite.mockPaymentRuleRepository.EXPECT().
				Create(gomock.Any(), gomock.Any()).
				DoAndReturn(func(_ context.Context, rule *models.PaymentRule) (*models.PaymentRule, error) {
					return rule, nil
				}).
				Times(2)
			suite.mockStrategy.EXPECT().SaveConfig(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil).Times(2)
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

	t.Run("Given the repository fails to persist the rule", func(t *testing.T) {
		t.Parallel()

		initSuite := func(suite *WhenCreatingPaymentRuleTestingSuite) {
			agent := model_builders.NewAgentBuilder().Build()
			policy := model_builders.NewPaymentPolicyBuilder().WithAgentId(agent.Id).Build()

			suite.mockStrategy.EXPECT().Validate(gomock.Any(), gomock.Any()).Return(nil).Times(1)
			suite.mockAgentRepository.EXPECT().Get(gomock.Any()).Return(agent, nil).Times(1)
			suite.mockPaymentPolicyRepository.EXPECT().Get(gomock.Any()).Return(policy, nil).Times(1)
			suite.mockPaymentRuleRepository.EXPECT().Create(gomock.Any(), gomock.Any()).Return(nil, assert.AnError).Times(1)
		}

		t.Run("Should return an internal error", func(t *testing.T) {
			t.Parallel()

			suite := WhenCreatingPaymentRuleBeforeEach(t)
			initSuite(suite)

			rule, err := suite.sut.Execute(context.Background(), generateSavePaymentRuleRequest())

			assert.Nil(t, rule)
			var internalErr *apperrors.InternalError
			assert.ErrorAs(t, err, &internalErr)
		})
	})

	t.Run("Given persisting the config fails", func(t *testing.T) {
		t.Parallel()

		initSuite := func(suite *WhenCreatingPaymentRuleTestingSuite) {
			agent := model_builders.NewAgentBuilder().Build()
			policy := model_builders.NewPaymentPolicyBuilder().WithAgentId(agent.Id).Build()
			rule := model_builders.NewPaymentRuleBuilder().WithPolicyId(policy.Id).Build()

			suite.mockStrategy.EXPECT().Validate(gomock.Any(), gomock.Any()).Return(nil).Times(1)
			suite.mockAgentRepository.EXPECT().Get(gomock.Any()).Return(agent, nil).Times(1)
			suite.mockPaymentPolicyRepository.EXPECT().Get(gomock.Any()).Return(policy, nil).Times(1)
			suite.mockPaymentRuleRepository.EXPECT().Create(gomock.Any(), gomock.Any()).Return(rule, nil).Times(1)
			suite.mockStrategy.EXPECT().SaveConfig(gomock.Any(), rule.Id, gomock.Any()).Return(assert.AnError).Times(1)
			// The orphaned rule is rolled back.
			suite.mockPaymentRuleRepository.EXPECT().DeleteById(gomock.Any(), rule.Id).Return(nil).Times(1)
		}

		t.Run("Should roll the rule back and return an internal error", func(t *testing.T) {
			t.Parallel()

			suite := WhenCreatingPaymentRuleBeforeEach(t)
			initSuite(suite)

			rule, err := suite.sut.Execute(context.Background(), generateSavePaymentRuleRequest())

			assert.Nil(t, rule)
			var internalErr *apperrors.InternalError
			assert.ErrorAs(t, err, &internalErr)
		})
	})

	t.Run("Given the strategy rejects the config", func(t *testing.T) {
		t.Parallel()

		t.Run("Should return the validation error before touching the policy", func(t *testing.T) {
			t.Parallel()

			suite := WhenCreatingPaymentRuleBeforeEach(t)
			suite.mockStrategy.EXPECT().
				Validate(gomock.Any(), gomock.Any()).
				Return(apperrors.NewValidationError("config.network", "unsupported")).
				Times(1)

			rule, err := suite.sut.Execute(context.Background(), generateSavePaymentRuleRequest())

			assert.Nil(t, rule)
			var validationErr *apperrors.ValidationError
			assert.ErrorAs(t, err, &validationErr)
			assert.Equal(t, "config.network", validationErr.Field)
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

	t.Run("Given the method is unsupported", func(t *testing.T) {
		t.Parallel()

		t.Run("Should return a validation error on the method field", func(t *testing.T) {
			t.Parallel()

			suite := WhenCreatingPaymentRuleBeforeEach(t)

			request := generateSavePaymentRuleRequest()
			request.Method = "lightning"
			rule, err := suite.sut.Execute(context.Background(), request)

			assert.Nil(t, rule)
			var validationErr *apperrors.ValidationError
			assert.ErrorAs(t, err, &validationErr)
		})
	})
}
