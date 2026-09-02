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
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

type WhenUpdatingPaymentSpendRuleTestingSuite struct {
	sut *use_cases.UpdatePaymentSpendRule

	mockPaymentSpendPolicyProvider *payment_interfaces_mocks.MockPaymentSpendPolicyProviderInterface
	mockPaymentSpendRuleRepository *payment_repositories_mocks.MockPaymentSpendRuleRepositoryInterface
	mockStrategy                   *payment_interfaces_mocks.MockPaymentSpendMethodStrategyInterface
}

func WhenUpdatingPaymentSpendRuleBeforeEach(t *testing.T) *WhenUpdatingPaymentSpendRuleTestingSuite {
	mockController := gomock.NewController(t)

	mockPaymentSpendPolicyProvider := payment_interfaces_mocks.NewMockPaymentSpendPolicyProviderInterface(
		mockController,
	)
	mockPaymentSpendRuleRepository := payment_repositories_mocks.NewMockPaymentSpendRuleRepositoryInterface(
		mockController,
	)
	strategies, mockStrategy := newMockSpendStrategies(mockController)

	return &WhenUpdatingPaymentSpendRuleTestingSuite{
		sut: use_cases.NewUpdatePaymentSpendRule(
			mockPaymentSpendPolicyProvider,
			mockPaymentSpendRuleRepository,
			strategies,
		),

		mockPaymentSpendPolicyProvider: mockPaymentSpendPolicyProvider,
		mockPaymentSpendRuleRepository: mockPaymentSpendRuleRepository,
		mockStrategy:                   mockStrategy,
	}
}

func generateUpdatePaymentSpendRuleRequest(t *testing.T, id uuid.UUID) *requests.UpdatePaymentSpendRuleRequest {
	t.Helper()

	return &requests.UpdatePaymentSpendRuleRequest{
		Id:                          id,
		SavePaymentSpendRuleRequest: *generateSavePaymentSpendRuleRequest(t),
	}
}

func TestWhenUpdatingPaymentSpendRule(t *testing.T) {
	t.Parallel()

	t.Run("Given the rule exists", func(t *testing.T) {
		t.Parallel()

		initSuite := func(suite *WhenUpdatingPaymentSpendRuleTestingSuite) *models.PaymentSpendRule {
			rule := model_builders.NewPaymentSpendRuleBuilder().Build()

			suite.mockStrategy.EXPECT().Validate(gomock.Any(), gomock.Any()).Return(nil).Times(1)
			suite.mockPaymentSpendRuleRepository.EXPECT().GetById(gomock.Any(), rule.Id).Return(rule, nil).Times(1)
			suite.mockPaymentSpendRuleRepository.EXPECT().
				Update(gomock.Any(), gomock.Any()).
				DoAndReturn(func(_ context.Context, updated *models.PaymentSpendRule) (*models.PaymentSpendRule, error) {
					assert.Equal(t, "Anything, unbounded", updated.Name)
					assert.Equal(t, "10000", updated.MaxAmount)
					assert.Equal(t, enums.PaymentMethodX402, updated.Method)
					return updated, nil
				}).
				Times(1)
			suite.mockStrategy.EXPECT().SaveConfig(gomock.Any(), rule.Id, gomock.Any()).Return(nil).Times(1)
			return rule
		}

		t.Run("Should replace the rule's fields and its matcher", func(t *testing.T) {
			t.Parallel()

			suite := WhenUpdatingPaymentSpendRuleBeforeEach(t)
			rule := initSuite(suite)

			updated, err := suite.sut.Execute(context.Background(), generateUpdatePaymentSpendRuleRequest(t, rule.Id))

			assert.NoError(t, err)
			assert.Equal(t, "Anything, unbounded", updated.Name)
		})
	})

	t.Run("Given the rule does not exist", func(t *testing.T) {
		t.Parallel()

		initSuite := func(suite *WhenUpdatingPaymentSpendRuleTestingSuite) {
			suite.mockStrategy.EXPECT().Validate(gomock.Any(), gomock.Any()).Return(nil).Times(1)
			suite.mockPaymentSpendRuleRepository.EXPECT().
				GetById(gomock.Any(), gomock.Any()).
				Return(nil, nil).
				Times(1)
		}

		t.Run("Should return a not found error", func(t *testing.T) {
			t.Parallel()

			suite := WhenUpdatingPaymentSpendRuleBeforeEach(t)
			initSuite(suite)

			updated, err := suite.sut.Execute(
				context.Background(),
				generateUpdatePaymentSpendRuleRequest(t, uuid.New()),
			)

			assert.Nil(t, updated)
			var notFoundErr *apperrors.NotFoundError
			assert.ErrorAs(t, err, &notFoundErr)
		})
	})

	t.Run("Given reading the rule fails", func(t *testing.T) {
		t.Parallel()

		initSuite := func(suite *WhenUpdatingPaymentSpendRuleTestingSuite) {
			suite.mockStrategy.EXPECT().Validate(gomock.Any(), gomock.Any()).Return(nil).Times(1)
			suite.mockPaymentSpendRuleRepository.EXPECT().
				GetById(gomock.Any(), gomock.Any()).
				Return(nil, assert.AnError).
				Times(1)
		}

		t.Run("Should return an internal error", func(t *testing.T) {
			t.Parallel()

			suite := WhenUpdatingPaymentSpendRuleBeforeEach(t)
			initSuite(suite)

			updated, err := suite.sut.Execute(
				context.Background(),
				generateUpdatePaymentSpendRuleRequest(t, uuid.New()),
			)

			assert.Nil(t, updated)
			var internalErr *apperrors.InternalError
			assert.ErrorAs(t, err, &internalErr)
		})
	})

	t.Run("Given persisting the rule fails", func(t *testing.T) {
		t.Parallel()

		initSuite := func(suite *WhenUpdatingPaymentSpendRuleTestingSuite) *models.PaymentSpendRule {
			rule := model_builders.NewPaymentSpendRuleBuilder().Build()

			suite.mockStrategy.EXPECT().Validate(gomock.Any(), gomock.Any()).Return(nil).Times(1)
			suite.mockPaymentSpendRuleRepository.EXPECT().GetById(gomock.Any(), rule.Id).Return(rule, nil).Times(1)
			suite.mockPaymentSpendRuleRepository.EXPECT().
				Update(gomock.Any(), gomock.Any()).
				Return(nil, assert.AnError).
				Times(1)
			return rule
		}

		t.Run("Should return an internal error", func(t *testing.T) {
			t.Parallel()

			suite := WhenUpdatingPaymentSpendRuleBeforeEach(t)
			rule := initSuite(suite)

			updated, err := suite.sut.Execute(context.Background(), generateUpdatePaymentSpendRuleRequest(t, rule.Id))

			assert.Nil(t, updated)
			var internalErr *apperrors.InternalError
			assert.ErrorAs(t, err, &internalErr)
		})
	})

	t.Run("Given persisting the matcher fails", func(t *testing.T) {
		t.Parallel()

		initSuite := func(suite *WhenUpdatingPaymentSpendRuleTestingSuite) *models.PaymentSpendRule {
			rule := model_builders.NewPaymentSpendRuleBuilder().Build()

			suite.mockStrategy.EXPECT().Validate(gomock.Any(), gomock.Any()).Return(nil).Times(1)
			suite.mockPaymentSpendRuleRepository.EXPECT().GetById(gomock.Any(), rule.Id).Return(rule, nil).Times(1)
			suite.mockPaymentSpendRuleRepository.EXPECT().Update(gomock.Any(), gomock.Any()).Return(rule, nil).Times(1)
			suite.mockStrategy.EXPECT().SaveConfig(gomock.Any(), rule.Id, gomock.Any()).Return(assert.AnError).Times(1)
			return rule
		}

		t.Run("Should return an internal error", func(t *testing.T) {
			t.Parallel()

			suite := WhenUpdatingPaymentSpendRuleBeforeEach(t)
			rule := initSuite(suite)

			updated, err := suite.sut.Execute(context.Background(), generateUpdatePaymentSpendRuleRequest(t, rule.Id))

			assert.Nil(t, updated)
			var internalErr *apperrors.InternalError
			assert.ErrorAs(t, err, &internalErr)
		})
	})

	t.Run("Given the strategy rejects the matcher", func(t *testing.T) {
		t.Parallel()

		initSuite := func(suite *WhenUpdatingPaymentSpendRuleTestingSuite) {
			suite.mockStrategy.EXPECT().
				Validate(gomock.Any(), gomock.Any()).
				Return(apperrors.NewValidationError("config.assetPatterns", "unsupported")).
				Times(1)
		}

		t.Run("Should return the validation error before reading the rule", func(t *testing.T) {
			t.Parallel()

			suite := WhenUpdatingPaymentSpendRuleBeforeEach(t)
			initSuite(suite)

			updated, err := suite.sut.Execute(
				context.Background(),
				generateUpdatePaymentSpendRuleRequest(t, uuid.New()),
			)

			assert.Nil(t, updated)
			var validationErr *apperrors.ValidationError
			assert.ErrorAs(t, err, &validationErr)
			assert.Equal(t, "config.assetPatterns", validationErr.Field)
		})
	})

	t.Run("Given the request is invalid", func(t *testing.T) {
		t.Parallel()

		t.Run("Should return a validation error before reading the rule", func(t *testing.T) {
			t.Parallel()

			suite := WhenUpdatingPaymentSpendRuleBeforeEach(t)

			request := generateUpdatePaymentSpendRuleRequest(t, uuid.New())
			request.MaxAmount = "-1"
			updated, err := suite.sut.Execute(context.Background(), request)

			assert.Nil(t, updated)
			var validationErr *apperrors.ValidationError
			assert.ErrorAs(t, err, &validationErr)
		})
	})

	t.Run("Given the method is unsupported", func(t *testing.T) {
		t.Parallel()

		t.Run("Should return a validation error on the method field", func(t *testing.T) {
			t.Parallel()

			suite := WhenUpdatingPaymentSpendRuleBeforeEach(t)

			request := generateUpdatePaymentSpendRuleRequest(t, uuid.New())
			request.Method = "lightning"
			updated, err := suite.sut.Execute(context.Background(), request)

			assert.Nil(t, updated)
			var validationErr *apperrors.ValidationError
			assert.ErrorAs(t, err, &validationErr)
		})
	})
}
