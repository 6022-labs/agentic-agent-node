package use_cases_test

import (
	"context"
	"encoding/json"
	"testing"

	identity_enums "github.com/6022-labs/agentic-agent-node/src/agent_identity/models/enums"
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

type WhenCreatingPaymentReceiveRuleTestingSuite struct {
	sut *use_cases.CreatePaymentReceiveRule

	mockAgentRepository                *identity_repositories_mocks.MockAgentRepositoryInterface
	mockAgentAddressRepository         *identity_repositories_mocks.MockAgentAddressRepositoryInterface
	mockPaymentReceivePolicyRepository *payment_repositories_mocks.MockPaymentReceivePolicyRepositoryInterface
	mockPaymentReceiveRuleRepository   *payment_repositories_mocks.MockPaymentReceiveRuleRepositoryInterface
	mockStrategy                       *payment_interfaces_mocks.MockPaymentReceiveMethodStrategyInterface
}

func WhenCreatingPaymentReceiveRuleBeforeEach(t *testing.T) *WhenCreatingPaymentReceiveRuleTestingSuite {
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

	return &WhenCreatingPaymentReceiveRuleTestingSuite{
		sut: use_cases.NewCreatePaymentReceiveRule(
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

func generateSavePaymentReceiveRuleRequest() *requests.SavePaymentReceiveRuleRequest {
	return &requests.SavePaymentReceiveRuleRequest{
		Name:  "Public price",
		Price: "10000",
		Config: json.RawMessage(
			`{"network":"eip155:80002","asset":"0x41E94Eb019C0762f9Bfcf9Fb1E58725BfB0e7582",` +
				`"payTo":"0xf39Fd6e51aad88F6F4ce6aB8827279cffFb92266","addressPatterns":[".*"]}`,
		),
	}
}

func TestWhenCreatingPaymentReceiveRule(t *testing.T) {
	t.Parallel()

	t.Run("Given an invalid request", func(t *testing.T) {
		t.Parallel()

		t.Run("Should return a validation error without touching dependencies", func(t *testing.T) {
			t.Parallel()

			suite := WhenCreatingPaymentReceiveRuleBeforeEach(t)

			request := generateSavePaymentReceiveRuleRequest()
			request.Name = ""
			_, err := suite.sut.Execute(context.Background(), request)

			var validationErr *apperrors.ValidationError
			assert.ErrorAs(t, err, &validationErr)
		})
	})

	t.Run("Given an unsupported method", func(t *testing.T) {
		t.Parallel()

		t.Run("Should return a validation error on the method field", func(t *testing.T) {
			t.Parallel()

			suite := WhenCreatingPaymentReceiveRuleBeforeEach(t)
			sut := use_cases.NewCreatePaymentReceiveRule(
				suite.mockAgentRepository,
				suite.mockAgentAddressRepository,
				suite.mockPaymentReceivePolicyRepository,
				suite.mockPaymentReceiveRuleRepository,
				newUnsupportedMethodStrategies(gomock.NewController(t)),
				&runtime_common_settings.Token6022Settings{},
			)

			rule, err := sut.Execute(context.Background(), generateSavePaymentReceiveRuleRequest())

			assert.Nil(t, rule)
			var validationErr *apperrors.ValidationError
			assert.ErrorAs(t, err, &validationErr)
			assert.Equal(t, "method", validationErr.Field)
		})
	})

	t.Run("Given the strategy rejects the config", func(t *testing.T) {
		t.Parallel()

		initSuite := func(suite *WhenCreatingPaymentReceiveRuleTestingSuite) {
			suite.mockStrategy.EXPECT().
				Validate(gomock.Any(), gomock.Any()).
				Return(apperrors.NewValidationError("config.network", "unsupported")).
				Times(1)
		}

		t.Run("Should return the validation error before touching the policy", func(t *testing.T) {
			t.Parallel()

			suite := WhenCreatingPaymentReceiveRuleBeforeEach(t)
			initSuite(suite)

			rule, err := suite.sut.Execute(context.Background(), generateSavePaymentReceiveRuleRequest())

			assert.Nil(t, rule)
			var validationErr *apperrors.ValidationError
			assert.ErrorAs(t, err, &validationErr)
			assert.Equal(t, "config.network", validationErr.Field)
		})
	})

	t.Run("Given no policy exists", func(t *testing.T) {
		t.Parallel()

		initSuite := func(suite *WhenCreatingPaymentReceiveRuleTestingSuite) {
			agent := model_builders.NewAgentBuilder().Build()
			policy := model_builders.NewPaymentReceivePolicyBuilder().WithAgentId(agent.Id).Build()

			suite.mockStrategy.EXPECT().Validate(gomock.Any(), gomock.Any()).Return(nil).Times(1)
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
			// One Create + SaveConfig seeds the default rule, the second pair persists the request's rule.
			suite.mockPaymentReceiveRuleRepository.EXPECT().
				Create(gomock.Any(), gomock.Any()).
				DoAndReturn(func(_ context.Context, rule *models.PaymentReceiveRule) (*models.PaymentReceiveRule, error) {
					return rule, nil
				}).
				Times(2)
			suite.mockStrategy.EXPECT().SaveConfig(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil).Times(2)
		}

		t.Run("Should materialize the policy before creating the rule", func(t *testing.T) {
			t.Parallel()

			suite := WhenCreatingPaymentReceiveRuleBeforeEach(t)
			initSuite(suite)

			rule, err := suite.sut.Execute(context.Background(), generateSavePaymentReceiveRuleRequest())

			assert.NoError(t, err)
			assert.Equal(t, "Public price", rule.Name)
		})
	})

	t.Run("Given saving the rule fails", func(t *testing.T) {
		t.Parallel()

		initSuite := func(suite *WhenCreatingPaymentReceiveRuleTestingSuite) {
			agent := model_builders.NewAgentBuilder().Build()
			policy := model_builders.NewPaymentReceivePolicyBuilder().WithAgentId(agent.Id).Build()

			suite.mockStrategy.EXPECT().Validate(gomock.Any(), gomock.Any()).Return(nil).Times(1)
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
				Create(gomock.Any(), gomock.Any()).
				Return(nil, assert.AnError).
				Times(1)
		}

		t.Run("Should return an InternalError", func(t *testing.T) {
			t.Parallel()

			suite := WhenCreatingPaymentReceiveRuleBeforeEach(t)
			initSuite(suite)

			rule, err := suite.sut.Execute(context.Background(), generateSavePaymentReceiveRuleRequest())

			assert.Nil(t, rule)
			var internalErr *apperrors.InternalError
			assert.ErrorAs(t, err, &internalErr)
		})
	})

	t.Run("Given saving the config fails", func(t *testing.T) {
		t.Parallel()

		initSuite := func(suite *WhenCreatingPaymentReceiveRuleTestingSuite) {
			agent := model_builders.NewAgentBuilder().Build()
			policy := model_builders.NewPaymentReceivePolicyBuilder().WithAgentId(agent.Id).Build()
			rule := model_builders.NewPaymentReceiveRuleBuilder().WithPolicyId(policy.Id).Build()

			suite.mockStrategy.EXPECT().Validate(gomock.Any(), gomock.Any()).Return(nil).Times(1)
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
				Create(gomock.Any(), gomock.Any()).
				Return(rule, nil).
				Times(1)
			suite.mockStrategy.EXPECT().SaveConfig(gomock.Any(), rule.Id, gomock.Any()).Return(assert.AnError).Times(1)
			// The orphaned rule is rolled back.
			suite.mockPaymentReceiveRuleRepository.EXPECT().DeleteById(gomock.Any(), rule.Id).Return(nil).Times(1)
		}

		t.Run("Should return an InternalError", func(t *testing.T) {
			t.Parallel()

			suite := WhenCreatingPaymentReceiveRuleBeforeEach(t)
			initSuite(suite)

			rule, err := suite.sut.Execute(context.Background(), generateSavePaymentReceiveRuleRequest())

			assert.Nil(t, rule)
			var internalErr *apperrors.InternalError
			assert.ErrorAs(t, err, &internalErr)
		})

		t.Run("Should roll the rule back", func(t *testing.T) {
			t.Parallel()

			suite := WhenCreatingPaymentReceiveRuleBeforeEach(t)
			initSuite(suite)

			_, _ = suite.sut.Execute(context.Background(), generateSavePaymentReceiveRuleRequest())
		})
	})

	t.Run("Given the rule config is valid", func(t *testing.T) {
		t.Parallel()

		initSuite := func(suite *WhenCreatingPaymentReceiveRuleTestingSuite) *models.PaymentReceivePolicy {
			agent := model_builders.NewAgentBuilder().Build()
			policy := model_builders.NewPaymentReceivePolicyBuilder().WithAgentId(agent.Id).Build()

			suite.mockStrategy.EXPECT().Validate(gomock.Any(), gomock.Any()).Return(nil).Times(1)
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
				Create(gomock.Any(), gomock.Any()).
				DoAndReturn(func(_ context.Context, rule *models.PaymentReceiveRule) (*models.PaymentReceiveRule, error) {
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

			suite := WhenCreatingPaymentReceiveRuleBeforeEach(t)
			initSuite(suite)

			rule, err := suite.sut.Execute(context.Background(), generateSavePaymentReceiveRuleRequest())

			assert.NoError(t, err)
			assert.Equal(t, "Public price", rule.Name)
		})
	})
}
