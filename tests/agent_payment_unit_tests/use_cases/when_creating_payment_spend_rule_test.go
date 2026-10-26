package use_cases_test

import (
	"context"
	"testing"

	"github.com/6022-labs/agentic-agent-node/src/agent_payment/models"
	"github.com/6022-labs/agentic-agent-node/src/agent_payment/models/enums"
	"github.com/6022-labs/agentic-agent-node/src/agent_payment/use_cases"
	"github.com/6022-labs/agentic-agent-node/src/agent_payment/use_cases/requests"
	apperrors "github.com/6022-labs/agentic-agent-node/src/common/errors"
	payment_repositories_mocks "github.com/6022-labs/agentic-agent-node/tests/agent_payment_mocks/repositories_mocks"
	payment_interfaces_mocks "github.com/6022-labs/agentic-agent-node/tests/agent_payment_mocks/services_mocks/interfaces_mocks"
	"github.com/6022-labs/agentic-agent-node/tests/common_tests/model_builders"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

type WhenCreatingPaymentSpendRuleTestingSuite struct {
	sut *use_cases.CreatePaymentSpendRule

	mockPaymentSpendPolicyProvider *payment_interfaces_mocks.MockPaymentSpendPolicyProviderInterface
	mockPaymentSpendRuleRepository *payment_repositories_mocks.MockPaymentSpendRuleRepositoryInterface
	mockStrategy                   *payment_interfaces_mocks.MockPaymentSpendMethodStrategyInterface
}

func WhenCreatingPaymentSpendRuleBeforeEach(t *testing.T) *WhenCreatingPaymentSpendRuleTestingSuite {
	mockController := gomock.NewController(t)

	mockPaymentSpendPolicyProvider := payment_interfaces_mocks.NewMockPaymentSpendPolicyProviderInterface(
		mockController,
	)
	mockPaymentSpendRuleRepository := payment_repositories_mocks.NewMockPaymentSpendRuleRepositoryInterface(
		mockController,
	)
	strategies, mockStrategy := newMockSpendStrategies(mockController)

	return &WhenCreatingPaymentSpendRuleTestingSuite{
		sut: use_cases.NewCreatePaymentSpendRule(
			mockPaymentSpendPolicyProvider,
			mockPaymentSpendRuleRepository,
			strategies,
		),

		mockPaymentSpendPolicyProvider: mockPaymentSpendPolicyProvider,
		mockPaymentSpendRuleRepository: mockPaymentSpendRuleRepository,
		mockStrategy:                   mockStrategy,
	}
}

func generateSavePaymentSpendRuleRequest(t *testing.T) *requests.SavePaymentSpendRuleRequest {
	t.Helper()

	config, err := model_builders.NewPaymentSpendRuleX402ConfigBuilder().Build().Marshal()
	assert.NoError(t, err)

	return &requests.SavePaymentSpendRuleRequest{
		Name:                "Anything, unbounded",
		MaxAmountPerPayment: "10000",
		Config:              config,
	}
}

func TestWhenCreatingPaymentSpendRule(t *testing.T) {
	t.Parallel()

	t.Run("Given the rule matcher is valid", func(t *testing.T) {
		t.Parallel()

		initSuite := func(suite *WhenCreatingPaymentSpendRuleTestingSuite) *models.PaymentSpendPolicy {
			policy := model_builders.NewPaymentSpendPolicyBuilder().Build()

			suite.mockStrategy.EXPECT().Validate(gomock.Any(), gomock.Any()).Return(nil).Times(1)
			suite.mockPaymentSpendPolicyProvider.EXPECT().Ensure(gomock.Any()).Return(policy, nil).Times(1)
			suite.mockPaymentSpendRuleRepository.EXPECT().
				Create(gomock.Any(), gomock.Any()).
				DoAndReturn(func(_ context.Context, rule *models.PaymentSpendRule) (*models.PaymentSpendRule, error) {
					assert.Equal(t, policy.Id, rule.PolicyId)
					assert.Equal(t, "Anything, unbounded", rule.Name)
					assert.Equal(t, "10000", rule.MaxAmountPerPayment)
					assert.Equal(t, enums.PaymentMethodX402, rule.Method)
					return rule, nil
				}).
				Times(1)
			// The matcher is persisted by the strategy, keyed by the rule.
			suite.mockStrategy.EXPECT().SaveConfig(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil).Times(1)
			return policy
		}

		t.Run("Should create the rule and persist its matcher", func(t *testing.T) {
			t.Parallel()

			suite := WhenCreatingPaymentSpendRuleBeforeEach(t)
			initSuite(suite)

			rule, err := suite.sut.Execute(context.Background(), generateSavePaymentSpendRuleRequest(t))

			assert.NoError(t, err)
			assert.Equal(t, "Anything, unbounded", rule.Name)
			assert.Equal(t, "10000", rule.MaxAmountPerPayment)
		})
	})

	t.Run("Given the rule carries a recurring budget", func(t *testing.T) {
		t.Parallel()

		initSuite := func(suite *WhenCreatingPaymentSpendRuleTestingSuite) {
			policy := model_builders.NewPaymentSpendPolicyBuilder().Build()

			suite.mockStrategy.EXPECT().Validate(gomock.Any(), gomock.Any()).Return(nil).Times(1)
			suite.mockPaymentSpendPolicyProvider.EXPECT().Ensure(gomock.Any()).Return(policy, nil).Times(1)
			suite.mockPaymentSpendRuleRepository.EXPECT().
				Create(gomock.Any(), gomock.Any()).
				DoAndReturn(func(_ context.Context, rule *models.PaymentSpendRule) (*models.PaymentSpendRule, error) {
					assert.Equal(t, 86400, rule.WindowSeconds)
					assert.Equal(t, "50000", rule.MaxAmountPerWindow)
					return rule, nil
				}).
				Times(1)
			suite.mockStrategy.EXPECT().SaveConfig(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil).Times(1)
		}

		t.Run("Should carry the period and its budget onto the rule", func(t *testing.T) {
			t.Parallel()

			suite := WhenCreatingPaymentSpendRuleBeforeEach(t)
			initSuite(suite)

			request := generateSavePaymentSpendRuleRequest(t)
			request.WindowSeconds = 86400
			request.MaxAmountPerWindow = "50000"

			rule, err := suite.sut.Execute(context.Background(), request)

			assert.NoError(t, err)
			assert.Equal(t, 86400, rule.WindowSeconds)
			assert.Equal(t, "50000", rule.MaxAmountPerWindow)
		})
	})

	t.Run("Given the repository fails to persist the rule", func(t *testing.T) {
		t.Parallel()

		initSuite := func(suite *WhenCreatingPaymentSpendRuleTestingSuite) {
			policy := model_builders.NewPaymentSpendPolicyBuilder().Build()

			suite.mockStrategy.EXPECT().Validate(gomock.Any(), gomock.Any()).Return(nil).Times(1)
			suite.mockPaymentSpendPolicyProvider.EXPECT().Ensure(gomock.Any()).Return(policy, nil).Times(1)
			suite.mockPaymentSpendRuleRepository.EXPECT().
				Create(gomock.Any(), gomock.Any()).
				Return(nil, assert.AnError).
				Times(1)
		}

		t.Run("Should return an internal error", func(t *testing.T) {
			t.Parallel()

			suite := WhenCreatingPaymentSpendRuleBeforeEach(t)
			initSuite(suite)

			rule, err := suite.sut.Execute(context.Background(), generateSavePaymentSpendRuleRequest(t))

			assert.Nil(t, rule)
			var internalErr *apperrors.InternalError
			assert.ErrorAs(t, err, &internalErr)
		})
	})

	t.Run("Given persisting the matcher fails", func(t *testing.T) {
		t.Parallel()

		initSuite := func(suite *WhenCreatingPaymentSpendRuleTestingSuite) {
			policy := model_builders.NewPaymentSpendPolicyBuilder().Build()
			rule := model_builders.NewPaymentSpendRuleBuilder().WithPolicyId(policy.Id).Build()

			suite.mockStrategy.EXPECT().Validate(gomock.Any(), gomock.Any()).Return(nil).Times(1)
			suite.mockPaymentSpendPolicyProvider.EXPECT().Ensure(gomock.Any()).Return(policy, nil).Times(1)
			suite.mockPaymentSpendRuleRepository.EXPECT().
				Create(gomock.Any(), gomock.Any()).
				Return(rule, nil).
				Times(1)
			suite.mockStrategy.EXPECT().SaveConfig(gomock.Any(), rule.Id, gomock.Any()).Return(assert.AnError).Times(1)
			// The orphaned rule is rolled back.
			suite.mockPaymentSpendRuleRepository.EXPECT().DeleteById(gomock.Any(), rule.Id).Return(nil).Times(1)
		}

		t.Run("Should roll the rule back and return an internal error", func(t *testing.T) {
			t.Parallel()

			suite := WhenCreatingPaymentSpendRuleBeforeEach(t)
			initSuite(suite)

			rule, err := suite.sut.Execute(context.Background(), generateSavePaymentSpendRuleRequest(t))

			assert.Nil(t, rule)
			var internalErr *apperrors.InternalError
			assert.ErrorAs(t, err, &internalErr)
		})
	})

	t.Run("Given the strategy rejects the matcher", func(t *testing.T) {
		t.Parallel()

		initSuite := func(suite *WhenCreatingPaymentSpendRuleTestingSuite) {
			suite.mockStrategy.EXPECT().
				Validate(gomock.Any(), gomock.Any()).
				Return(apperrors.NewValidationError("config.network", "unsupported")).
				Times(1)
		}

		t.Run("Should return the validation error before touching the policy", func(t *testing.T) {
			t.Parallel()

			suite := WhenCreatingPaymentSpendRuleBeforeEach(t)
			initSuite(suite)

			rule, err := suite.sut.Execute(context.Background(), generateSavePaymentSpendRuleRequest(t))

			assert.Nil(t, rule)
			var validationErr *apperrors.ValidationError
			assert.ErrorAs(t, err, &validationErr)
			assert.Equal(t, "config.network", validationErr.Field)
		})
	})

	t.Run("Given materializing the policy fails", func(t *testing.T) {
		t.Parallel()

		initSuite := func(suite *WhenCreatingPaymentSpendRuleTestingSuite) {
			suite.mockStrategy.EXPECT().Validate(gomock.Any(), gomock.Any()).Return(nil).Times(1)
			suite.mockPaymentSpendPolicyProvider.EXPECT().
				Ensure(gomock.Any()).
				Return(nil, apperrors.NewInternalError("ensure_failed", "")).
				Times(1)
		}

		t.Run("Should propagate the provider's error", func(t *testing.T) {
			t.Parallel()

			suite := WhenCreatingPaymentSpendRuleBeforeEach(t)
			initSuite(suite)

			rule, err := suite.sut.Execute(context.Background(), generateSavePaymentSpendRuleRequest(t))

			assert.Nil(t, rule)
			var internalErr *apperrors.InternalError
			assert.ErrorAs(t, err, &internalErr)
		})
	})

	t.Run("Given the request is invalid", func(t *testing.T) {
		t.Parallel()

		t.Run("Should return a validation error before touching the policy", func(t *testing.T) {
			t.Parallel()

			suite := WhenCreatingPaymentSpendRuleBeforeEach(t)

			request := generateSavePaymentSpendRuleRequest(t)
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

			suite := WhenCreatingPaymentSpendRuleBeforeEach(t)

			request := generateSavePaymentSpendRuleRequest(t)
			request.Method = "lightning"
			rule, err := suite.sut.Execute(context.Background(), request)

			assert.Nil(t, rule)
			var validationErr *apperrors.ValidationError
			assert.ErrorAs(t, err, &validationErr)
		})
	})
}
