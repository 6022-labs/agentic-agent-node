package use_cases_test

import (
	"context"
	"testing"
	"time"

	"github.com/6022-labs/agentic-agent-node/src/agent_payment/models"
	"github.com/6022-labs/agentic-agent-node/src/agent_payment/use_cases"
	"github.com/6022-labs/agentic-agent-node/src/agent_payment/use_cases/requests"
	apperrors "github.com/6022-labs/agentic-agent-node/src/common/errors"
	payment_repositories_mocks "github.com/6022-labs/agentic-agent-node/tests/agent_payment_mocks/repositories_mocks"
	payment_interfaces_mocks "github.com/6022-labs/agentic-agent-node/tests/agent_payment_mocks/services_mocks/interfaces_mocks"
	"github.com/6022-labs/agentic-agent-node/tests/common_tests/model_builders"
	runtime_common_mocks "github.com/6022-labs/agentic-agent-node/tests/runtime_common_mocks/services_mocks/interfaces_mocks"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

type WhenUpdatingPaymentSpendRuleTestingSuite struct {
	sut *use_cases.UpdatePaymentSpendRule

	mockPaymentSpendRuleRepository       *payment_repositories_mocks.MockPaymentSpendRuleRepositoryInterface
	mockPaymentSpendRuleWindowRepository *payment_repositories_mocks.MockPaymentSpendRuleWindowRepositoryInterface
	mockStrategy                         *payment_interfaces_mocks.MockPaymentSpendMethodStrategyInterface
	mockUnitOfWorkFactory                *runtime_common_mocks.MockUnitOfWorkFactoryInterface
	mockUnitOfWork                       *runtime_common_mocks.MockUnitOfWorkInterface
	unitOfWorkCtx                        context.Context
}

func WhenUpdatingPaymentSpendRuleBeforeEach(t *testing.T) *WhenUpdatingPaymentSpendRuleTestingSuite {
	mockController := gomock.NewController(t)

	mockPaymentSpendRuleRepository := payment_repositories_mocks.NewMockPaymentSpendRuleRepositoryInterface(
		mockController,
	)
	mockPaymentSpendRuleWindowRepository :=
		payment_repositories_mocks.NewMockPaymentSpendRuleWindowRepositoryInterface(mockController)
	strategies, mockStrategy := newMockSpendStrategies(mockController)
	mockUnitOfWorkFactory := runtime_common_mocks.NewMockUnitOfWorkFactoryInterface(mockController)
	mockUnitOfWork, unitOfWorkCtx := newMockUnitOfWork(mockController)

	return &WhenUpdatingPaymentSpendRuleTestingSuite{
		sut: use_cases.NewUpdatePaymentSpendRule(
			payment_interfaces_mocks.NewMockPaymentSpendPolicyProviderInterface(mockController),
			mockPaymentSpendRuleRepository,
			mockPaymentSpendRuleWindowRepository,
			strategies,
			mockUnitOfWorkFactory,
		),

		mockPaymentSpendRuleRepository:       mockPaymentSpendRuleRepository,
		mockPaymentSpendRuleWindowRepository: mockPaymentSpendRuleWindowRepository,
		mockStrategy:                         mockStrategy,
		mockUnitOfWorkFactory:                mockUnitOfWorkFactory,
		mockUnitOfWork:                       mockUnitOfWork,
		unitOfWorkCtx:                        unitOfWorkCtx,
	}
}

func generateUpdatePaymentSpendRuleRequest(t *testing.T, id uuid.UUID) *requests.UpdatePaymentSpendRuleRequest {
	t.Helper()

	return &requests.UpdatePaymentSpendRuleRequest{
		Id:                          id,
		SavePaymentSpendRuleRequest: *generateSavePaymentSpendRuleRequest(t),
	}
}

func generateWindowedUpdateRequest(t *testing.T, id uuid.UUID) *requests.UpdatePaymentSpendRuleRequest {
	t.Helper()

	request := generateUpdatePaymentSpendRuleRequest(t, id)
	request.Windows = []requests.SavePaymentSpendRuleWindowRequest{
		{WindowSeconds: 86400, MaxAmount: "60000"},
		{WindowSeconds: 604800, MaxAmount: "200000"},
	}
	return request
}

func TestWhenUpdatingPaymentSpendRule(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	t.Run("Given the request is invalid", func(t *testing.T) {
		t.Parallel()

		t.Run("Should return a validation error before reading the rule", func(t *testing.T) {
			t.Parallel()

			suite := WhenUpdatingPaymentSpendRuleBeforeEach(t)
			request := generateUpdatePaymentSpendRuleRequest(t, uuid.New())
			request.Name = ""

			rule, err := suite.sut.Execute(ctx, request)

			assert.Nil(t, rule)
			var validationErr *apperrors.ValidationError
			assert.ErrorAs(t, err, &validationErr)
			assert.Equal(t, "name", validationErr.Field)
		})
	})

	t.Run("Given the method is unsupported", func(t *testing.T) {
		t.Parallel()

		t.Run("Should return a validation error on the method field", func(t *testing.T) {
			t.Parallel()

			suite := WhenUpdatingPaymentSpendRuleBeforeEach(t)
			sut := use_cases.NewUpdatePaymentSpendRule(
				payment_interfaces_mocks.NewMockPaymentSpendPolicyProviderInterface(gomock.NewController(t)),
				suite.mockPaymentSpendRuleRepository,
				suite.mockPaymentSpendRuleWindowRepository,
				newUnsupportedSpendMethodStrategies(gomock.NewController(t)),
				suite.mockUnitOfWorkFactory,
			)

			rule, err := sut.Execute(ctx, generateUpdatePaymentSpendRuleRequest(t, uuid.New()))

			assert.Nil(t, rule)
			var validationErr *apperrors.ValidationError
			assert.ErrorAs(t, err, &validationErr)
			assert.Equal(t, "method", validationErr.Field)
		})
	})

	t.Run("Given the strategy rejects the token", func(t *testing.T) {
		t.Parallel()

		initSuite := func(suite *WhenUpdatingPaymentSpendRuleTestingSuite) {
			suite.mockStrategy.EXPECT().
				Validate(gomock.Any(), gomock.Any()).
				Return(apperrors.NewValidationError("config.network", "unsupported")).
				Times(1)
		}

		t.Run("Should return a validation error before reading the rule", func(t *testing.T) {
			t.Parallel()

			suite := WhenUpdatingPaymentSpendRuleBeforeEach(t)
			initSuite(suite)

			rule, err := suite.sut.Execute(ctx, generateUpdatePaymentSpendRuleRequest(t, uuid.New()))

			assert.Nil(t, rule)
			var validationErr *apperrors.ValidationError
			assert.ErrorAs(t, err, &validationErr)
			assert.Equal(t, "config.network", validationErr.Field)
		})
	})

	t.Run("Given reading the rule fails", func(t *testing.T) {
		t.Parallel()

		ruleId := uuid.New()

		initSuite := func(suite *WhenUpdatingPaymentSpendRuleTestingSuite) {
			suite.mockStrategy.EXPECT().Validate(gomock.Any(), gomock.Any()).Return(nil).Times(1)
			suite.mockPaymentSpendRuleRepository.EXPECT().
				GetById(gomock.Any(), ruleId).
				Return(nil, assert.AnError).
				Times(1)
		}

		t.Run("Should return an InternalError", func(t *testing.T) {
			t.Parallel()

			suite := WhenUpdatingPaymentSpendRuleBeforeEach(t)
			initSuite(suite)

			rule, err := suite.sut.Execute(ctx, generateUpdatePaymentSpendRuleRequest(t, ruleId))

			assert.Nil(t, rule)
			var internalErr *apperrors.InternalError
			assert.ErrorAs(t, err, &internalErr)
			assert.Equal(t, "update_failed", internalErr.Code)
		})
	})

	t.Run("Given the rule does not exist", func(t *testing.T) {
		t.Parallel()

		ruleId := uuid.New()

		initSuite := func(suite *WhenUpdatingPaymentSpendRuleTestingSuite) {
			suite.mockStrategy.EXPECT().Validate(gomock.Any(), gomock.Any()).Return(nil).Times(1)
			suite.mockPaymentSpendRuleRepository.EXPECT().GetById(gomock.Any(), ruleId).Return(nil, nil).Times(1)
		}

		t.Run("Should return a not found error", func(t *testing.T) {
			t.Parallel()

			suite := WhenUpdatingPaymentSpendRuleBeforeEach(t)
			initSuite(suite)

			rule, err := suite.sut.Execute(ctx, generateUpdatePaymentSpendRuleRequest(t, ruleId))

			assert.Nil(t, rule)
			var notFoundErr *apperrors.NotFoundError
			assert.ErrorAs(t, err, &notFoundErr)
			assert.Equal(t, "rule_not_found", notFoundErr.Code)
		})
	})

	t.Run("Given looking up the token's rule fails", func(t *testing.T) {
		t.Parallel()

		initSuite := func(suite *WhenUpdatingPaymentSpendRuleTestingSuite) *models.PaymentSpendRule {
			rule := model_builders.NewPaymentSpendRuleBuilder().Build()

			suite.mockStrategy.EXPECT().Validate(gomock.Any(), gomock.Any()).Return(nil).Times(1)
			suite.mockPaymentSpendRuleRepository.EXPECT().GetById(gomock.Any(), rule.Id).Return(rule, nil).Times(1)
			suite.mockStrategy.EXPECT().
				RuleIdForAsset(gomock.Any(), rule.PolicyId, gomock.Any()).
				Return(uuid.Nil, assert.AnError).
				Times(1)
			return rule
		}

		t.Run("Should return an InternalError", func(t *testing.T) {
			t.Parallel()

			suite := WhenUpdatingPaymentSpendRuleBeforeEach(t)
			rule := initSuite(suite)

			response, err := suite.sut.Execute(ctx, generateUpdatePaymentSpendRuleRequest(t, rule.Id))

			assert.Nil(t, response)
			var internalErr *apperrors.InternalError
			assert.ErrorAs(t, err, &internalErr)
			assert.Equal(t, "rule_lookup_failed", internalErr.Code)
		})
	})

	t.Run("Given another rule already spends the token", func(t *testing.T) {
		t.Parallel()

		initSuite := func(suite *WhenUpdatingPaymentSpendRuleTestingSuite) *models.PaymentSpendRule {
			rule := model_builders.NewPaymentSpendRuleBuilder().Build()

			suite.mockStrategy.EXPECT().Validate(gomock.Any(), gomock.Any()).Return(nil).Times(1)
			suite.mockPaymentSpendRuleRepository.EXPECT().GetById(gomock.Any(), rule.Id).Return(rule, nil).Times(1)
			suite.mockStrategy.EXPECT().
				RuleIdForAsset(gomock.Any(), rule.PolicyId, gomock.Any()).
				Return(uuid.New(), nil).
				Times(1)
			return rule
		}

		t.Run("Should return a conflict error", func(t *testing.T) {
			t.Parallel()

			suite := WhenUpdatingPaymentSpendRuleBeforeEach(t)
			rule := initSuite(suite)

			response, err := suite.sut.Execute(ctx, generateUpdatePaymentSpendRuleRequest(t, rule.Id))

			assert.Nil(t, response)
			var conflictErr *apperrors.ConflictError
			assert.ErrorAs(t, err, &conflictErr)
			assert.Equal(t, "asset_already_has_rule", conflictErr.Code)
		})
	})

	t.Run("Given starting the unit of work fails", func(t *testing.T) {
		t.Parallel()

		initSuite := func(suite *WhenUpdatingPaymentSpendRuleTestingSuite) *models.PaymentSpendRule {
			rule := model_builders.NewPaymentSpendRuleBuilder().Build()

			suite.mockStrategy.EXPECT().Validate(gomock.Any(), gomock.Any()).Return(nil).Times(1)
			suite.mockPaymentSpendRuleRepository.EXPECT().GetById(gomock.Any(), rule.Id).Return(rule, nil).Times(1)
			suite.mockStrategy.EXPECT().
				RuleIdForAsset(gomock.Any(), rule.PolicyId, gomock.Any()).
				Return(rule.Id, nil).
				Times(1)
			suite.mockUnitOfWorkFactory.EXPECT().Start(gomock.Any()).Return(nil, assert.AnError).Times(1)
			return rule
		}

		t.Run("Should return an InternalError", func(t *testing.T) {
			t.Parallel()

			suite := WhenUpdatingPaymentSpendRuleBeforeEach(t)
			rule := initSuite(suite)

			response, err := suite.sut.Execute(ctx, generateUpdatePaymentSpendRuleRequest(t, rule.Id))

			assert.Nil(t, response)
			var internalErr *apperrors.InternalError
			assert.ErrorAs(t, err, &internalErr)
			assert.Equal(t, "update_failed", internalErr.Code)
		})
	})

	t.Run("Given the repository fails to persist the rule", func(t *testing.T) {
		t.Parallel()

		initSuite := func(suite *WhenUpdatingPaymentSpendRuleTestingSuite) *models.PaymentSpendRule {
			rule := model_builders.NewPaymentSpendRuleBuilder().Build()

			suite.mockStrategy.EXPECT().Validate(gomock.Any(), gomock.Any()).Return(nil).Times(1)
			suite.mockPaymentSpendRuleRepository.EXPECT().GetById(gomock.Any(), rule.Id).Return(rule, nil).Times(1)
			suite.mockStrategy.EXPECT().
				RuleIdForAsset(gomock.Any(), rule.PolicyId, gomock.Any()).
				Return(rule.Id, nil).
				Times(1)
			suite.mockUnitOfWorkFactory.EXPECT().Start(gomock.Any()).Return(suite.mockUnitOfWork, nil).Times(1)
			suite.mockPaymentSpendRuleRepository.EXPECT().
				Update(gomock.Any(), gomock.Any()).
				Return(nil, assert.AnError).
				Times(1)
			return rule
		}

		t.Run("Should return an InternalError", func(t *testing.T) {
			t.Parallel()

			suite := WhenUpdatingPaymentSpendRuleBeforeEach(t)
			rule := initSuite(suite)

			response, err := suite.sut.Execute(ctx, generateUpdatePaymentSpendRuleRequest(t, rule.Id))

			assert.Nil(t, response)
			var internalErr *apperrors.InternalError
			assert.ErrorAs(t, err, &internalErr)
			assert.Equal(t, "update_failed", internalErr.Code)
		})
	})

	t.Run("Given persisting the token fails", func(t *testing.T) {
		t.Parallel()

		initSuite := func(suite *WhenUpdatingPaymentSpendRuleTestingSuite) *models.PaymentSpendRule {
			rule := model_builders.NewPaymentSpendRuleBuilder().Build()

			suite.mockStrategy.EXPECT().Validate(gomock.Any(), gomock.Any()).Return(nil).Times(1)
			suite.mockPaymentSpendRuleRepository.EXPECT().GetById(gomock.Any(), rule.Id).Return(rule, nil).Times(1)
			suite.mockStrategy.EXPECT().
				RuleIdForAsset(gomock.Any(), rule.PolicyId, gomock.Any()).
				Return(rule.Id, nil).
				Times(1)
			suite.mockUnitOfWorkFactory.EXPECT().Start(gomock.Any()).Return(suite.mockUnitOfWork, nil).Times(1)
			suite.mockPaymentSpendRuleRepository.EXPECT().Update(gomock.Any(), gomock.Any()).Return(rule, nil).Times(1)
			suite.mockStrategy.EXPECT().
				SaveConfig(gomock.Any(), rule.PolicyId, rule.Id, gomock.Any()).
				Return(assert.AnError).
				Times(1)
			return rule
		}

		t.Run("Should return an InternalError", func(t *testing.T) {
			t.Parallel()

			suite := WhenUpdatingPaymentSpendRuleBeforeEach(t)
			rule := initSuite(suite)

			response, err := suite.sut.Execute(ctx, generateUpdatePaymentSpendRuleRequest(t, rule.Id))

			assert.Nil(t, response)
			var internalErr *apperrors.InternalError
			assert.ErrorAs(t, err, &internalErr)
			assert.Equal(t, "update_failed", internalErr.Code)
		})
	})

	t.Run("Given reading the rule's windows fails", func(t *testing.T) {
		t.Parallel()

		initSuite := func(suite *WhenUpdatingPaymentSpendRuleTestingSuite) *models.PaymentSpendRule {
			rule := model_builders.NewPaymentSpendRuleBuilder().Build()

			suite.mockStrategy.EXPECT().Validate(gomock.Any(), gomock.Any()).Return(nil).Times(1)
			suite.mockPaymentSpendRuleRepository.EXPECT().GetById(gomock.Any(), rule.Id).Return(rule, nil).Times(1)
			suite.mockStrategy.EXPECT().
				RuleIdForAsset(gomock.Any(), rule.PolicyId, gomock.Any()).
				Return(rule.Id, nil).
				Times(1)
			suite.mockUnitOfWorkFactory.EXPECT().Start(gomock.Any()).Return(suite.mockUnitOfWork, nil).Times(1)
			suite.mockPaymentSpendRuleRepository.EXPECT().Update(gomock.Any(), gomock.Any()).Return(rule, nil).Times(1)
			suite.mockStrategy.EXPECT().
				SaveConfig(gomock.Any(), rule.PolicyId, rule.Id, gomock.Any()).
				Return(nil).
				Times(1)
			suite.mockPaymentSpendRuleWindowRepository.EXPECT().
				ListByRuleIdForUpdate(gomock.Any(), rule.Id).
				Return(nil, assert.AnError).
				Times(1)
			return rule
		}

		t.Run("Should return an InternalError", func(t *testing.T) {
			t.Parallel()

			suite := WhenUpdatingPaymentSpendRuleBeforeEach(t)
			rule := initSuite(suite)

			response, err := suite.sut.Execute(ctx, generateUpdatePaymentSpendRuleRequest(t, rule.Id))

			assert.Nil(t, response)
			var internalErr *apperrors.InternalError
			assert.ErrorAs(t, err, &internalErr)
			assert.Equal(t, "update_failed", internalErr.Code)
		})
	})

	t.Run("Given persisting a window fails", func(t *testing.T) {
		t.Parallel()

		initSuite := func(suite *WhenUpdatingPaymentSpendRuleTestingSuite) *models.PaymentSpendRule {
			rule := model_builders.NewPaymentSpendRuleBuilder().Build()

			suite.mockStrategy.EXPECT().Validate(gomock.Any(), gomock.Any()).Return(nil).Times(1)
			suite.mockPaymentSpendRuleRepository.EXPECT().GetById(gomock.Any(), rule.Id).Return(rule, nil).Times(1)
			suite.mockStrategy.EXPECT().
				RuleIdForAsset(gomock.Any(), rule.PolicyId, gomock.Any()).
				Return(rule.Id, nil).
				Times(1)
			suite.mockUnitOfWorkFactory.EXPECT().Start(gomock.Any()).Return(suite.mockUnitOfWork, nil).Times(1)
			suite.mockPaymentSpendRuleRepository.EXPECT().Update(gomock.Any(), gomock.Any()).Return(rule, nil).Times(1)
			suite.mockStrategy.EXPECT().
				SaveConfig(gomock.Any(), rule.PolicyId, rule.Id, gomock.Any()).
				Return(nil).
				Times(1)
			suite.mockPaymentSpendRuleWindowRepository.EXPECT().
				ListByRuleIdForUpdate(gomock.Any(), rule.Id).
				Return([]models.PaymentSpendRuleWindow{}, nil).
				Times(1)
			suite.mockPaymentSpendRuleWindowRepository.EXPECT().
				Create(gomock.Any(), gomock.Any()).
				Return(nil, assert.AnError).
				Times(1)
			return rule
		}

		t.Run("Should return an InternalError", func(t *testing.T) {
			t.Parallel()

			suite := WhenUpdatingPaymentSpendRuleBeforeEach(t)
			rule := initSuite(suite)

			response, err := suite.sut.Execute(ctx, generateUpdatePaymentSpendRuleRequest(t, rule.Id))

			assert.Nil(t, response)
			var internalErr *apperrors.InternalError
			assert.ErrorAs(t, err, &internalErr)
			assert.Equal(t, "update_failed", internalErr.Code)
		})
	})

	t.Run("Given persisting a kept window fails", func(t *testing.T) {
		t.Parallel()

		initSuite := func(suite *WhenUpdatingPaymentSpendRuleTestingSuite) *models.PaymentSpendRule {
			rule := model_builders.NewPaymentSpendRuleBuilder().Build()
			daily := model_builders.NewPaymentSpendRuleWindowBuilder().WithPaymentSpendRuleId(rule.Id).Build()

			suite.mockStrategy.EXPECT().Validate(gomock.Any(), gomock.Any()).Return(nil).Times(1)
			suite.mockPaymentSpendRuleRepository.EXPECT().GetById(gomock.Any(), rule.Id).Return(rule, nil).Times(1)
			suite.mockStrategy.EXPECT().
				RuleIdForAsset(gomock.Any(), rule.PolicyId, gomock.Any()).
				Return(rule.Id, nil).
				Times(1)
			suite.mockUnitOfWorkFactory.EXPECT().Start(gomock.Any()).Return(suite.mockUnitOfWork, nil).Times(1)
			suite.mockPaymentSpendRuleRepository.EXPECT().Update(gomock.Any(), gomock.Any()).Return(rule, nil).Times(1)
			suite.mockStrategy.EXPECT().
				SaveConfig(gomock.Any(), rule.PolicyId, rule.Id, gomock.Any()).
				Return(nil).
				Times(1)
			suite.mockPaymentSpendRuleWindowRepository.EXPECT().
				ListByRuleIdForUpdate(gomock.Any(), rule.Id).
				Return([]models.PaymentSpendRuleWindow{*daily}, nil).
				Times(1)
			suite.mockPaymentSpendRuleWindowRepository.EXPECT().
				Update(gomock.Any(), gomock.Any()).
				Return(nil, assert.AnError).
				Times(1)
			return rule
		}

		t.Run("Should return an InternalError", func(t *testing.T) {
			t.Parallel()

			suite := WhenUpdatingPaymentSpendRuleBeforeEach(t)
			rule := initSuite(suite)

			response, err := suite.sut.Execute(ctx, generateUpdatePaymentSpendRuleRequest(t, rule.Id))

			assert.Nil(t, response)
			var internalErr *apperrors.InternalError
			assert.ErrorAs(t, err, &internalErr)
			assert.Equal(t, "update_failed", internalErr.Code)
		})
	})

	t.Run("Given removing a window fails", func(t *testing.T) {
		t.Parallel()

		initSuite := func(suite *WhenUpdatingPaymentSpendRuleTestingSuite) *models.PaymentSpendRule {
			rule := model_builders.NewPaymentSpendRuleBuilder().Build()
			dropped := model_builders.NewPaymentSpendRuleWindowBuilder().
				WithPaymentSpendRuleId(rule.Id).
				WithWindowSeconds(3600).
				Build()

			suite.mockStrategy.EXPECT().Validate(gomock.Any(), gomock.Any()).Return(nil).Times(1)
			suite.mockPaymentSpendRuleRepository.EXPECT().GetById(gomock.Any(), rule.Id).Return(rule, nil).Times(1)
			suite.mockStrategy.EXPECT().
				RuleIdForAsset(gomock.Any(), rule.PolicyId, gomock.Any()).
				Return(rule.Id, nil).
				Times(1)
			suite.mockUnitOfWorkFactory.EXPECT().Start(gomock.Any()).Return(suite.mockUnitOfWork, nil).Times(1)
			suite.mockPaymentSpendRuleRepository.EXPECT().Update(gomock.Any(), gomock.Any()).Return(rule, nil).Times(1)
			suite.mockStrategy.EXPECT().
				SaveConfig(gomock.Any(), rule.PolicyId, rule.Id, gomock.Any()).
				Return(nil).
				Times(1)
			suite.mockPaymentSpendRuleWindowRepository.EXPECT().
				ListByRuleIdForUpdate(gomock.Any(), rule.Id).
				Return([]models.PaymentSpendRuleWindow{*dropped}, nil).
				Times(1)
			suite.mockPaymentSpendRuleWindowRepository.EXPECT().
				Create(gomock.Any(), gomock.Any()).
				DoAndReturn(func(_ context.Context, window *models.PaymentSpendRuleWindow) (*models.PaymentSpendRuleWindow, error) {
					return window, nil
				}).
				Times(1)
			suite.mockPaymentSpendRuleWindowRepository.EXPECT().
				DeleteById(gomock.Any(), dropped.Id).
				Return(assert.AnError).
				Times(1)
			return rule
		}

		t.Run("Should return an InternalError", func(t *testing.T) {
			t.Parallel()

			suite := WhenUpdatingPaymentSpendRuleBeforeEach(t)
			rule := initSuite(suite)

			response, err := suite.sut.Execute(ctx, generateUpdatePaymentSpendRuleRequest(t, rule.Id))

			assert.Nil(t, response)
			var internalErr *apperrors.InternalError
			assert.ErrorAs(t, err, &internalErr)
			assert.Equal(t, "update_failed", internalErr.Code)
		})
	})

	t.Run("Given committing the rule fails", func(t *testing.T) {
		t.Parallel()

		initSuite := func(suite *WhenUpdatingPaymentSpendRuleTestingSuite) *models.PaymentSpendRule {
			rule := model_builders.NewPaymentSpendRuleBuilder().Build()

			suite.mockStrategy.EXPECT().Validate(gomock.Any(), gomock.Any()).Return(nil).Times(1)
			suite.mockPaymentSpendRuleRepository.EXPECT().GetById(gomock.Any(), rule.Id).Return(rule, nil).Times(1)
			suite.mockStrategy.EXPECT().
				RuleIdForAsset(gomock.Any(), rule.PolicyId, gomock.Any()).
				Return(rule.Id, nil).
				Times(1)
			suite.mockUnitOfWorkFactory.EXPECT().Start(gomock.Any()).Return(suite.mockUnitOfWork, nil).Times(1)
			suite.mockPaymentSpendRuleRepository.EXPECT().Update(gomock.Any(), gomock.Any()).Return(rule, nil).Times(1)
			suite.mockStrategy.EXPECT().
				SaveConfig(gomock.Any(), rule.PolicyId, rule.Id, gomock.Any()).
				Return(nil).
				Times(1)
			suite.mockPaymentSpendRuleWindowRepository.EXPECT().
				ListByRuleIdForUpdate(gomock.Any(), rule.Id).
				Return([]models.PaymentSpendRuleWindow{}, nil).
				Times(1)
			suite.mockPaymentSpendRuleWindowRepository.EXPECT().
				Create(gomock.Any(), gomock.Any()).
				DoAndReturn(func(_ context.Context, window *models.PaymentSpendRuleWindow) (*models.PaymentSpendRuleWindow, error) {
					return window, nil
				}).
				Times(1)
			suite.mockUnitOfWork.EXPECT().Commit().Return(assert.AnError).Times(1)
			return rule
		}

		t.Run("Should return an InternalError", func(t *testing.T) {
			t.Parallel()

			suite := WhenUpdatingPaymentSpendRuleBeforeEach(t)
			rule := initSuite(suite)

			response, err := suite.sut.Execute(ctx, generateUpdatePaymentSpendRuleRequest(t, rule.Id))

			assert.Nil(t, response)
			var internalErr *apperrors.InternalError
			assert.ErrorAs(t, err, &internalErr)
			assert.Equal(t, "update_failed", internalErr.Code)
		})
	})

	t.Run("Given the rule keeps its token", func(t *testing.T) {
		t.Parallel()

		// The daily window has spend recorded; the hourly one is no longer requested.
		initSuite := func(
			suite *WhenUpdatingPaymentSpendRuleTestingSuite,
		) (*models.PaymentSpendRule, *models.PaymentSpendRuleWindow, *models.PaymentSpendRuleWindow) {
			rule := model_builders.NewPaymentSpendRuleBuilder().WithName("Old name").Build()
			daily := model_builders.NewPaymentSpendRuleWindowBuilder().
				WithPaymentSpendRuleId(rule.Id).
				WithMaxAmount("50000").
				WithSpent("4000").
				WithStartedAt(time.Now().Add(-time.Hour)).
				Build()
			hourly := model_builders.NewPaymentSpendRuleWindowBuilder().
				WithPaymentSpendRuleId(rule.Id).
				WithWindowSeconds(3600).
				Build()

			suite.mockStrategy.EXPECT().Validate(ctx, gomock.Any()).Return(nil).Times(1)
			suite.mockPaymentSpendRuleRepository.EXPECT().GetById(ctx, rule.Id).Return(rule, nil).Times(1)
			suite.mockStrategy.EXPECT().
				RuleIdForAsset(ctx, rule.PolicyId, gomock.Any()).
				Return(rule.Id, nil).
				Times(1)
			suite.mockUnitOfWorkFactory.EXPECT().Start(ctx).Return(suite.mockUnitOfWork, nil).Times(1)
			suite.mockPaymentSpendRuleWindowRepository.EXPECT().
				ListByRuleIdForUpdate(suite.unitOfWorkCtx, rule.Id).
				Return([]models.PaymentSpendRuleWindow{*daily, *hourly}, nil).
				Times(1)
			suite.mockUnitOfWork.EXPECT().Commit().Return(nil).Times(1)
			return rule, daily, hourly
		}

		t.Run("Should replace the rule's fields", func(t *testing.T) {
			t.Parallel()

			suite := WhenUpdatingPaymentSpendRuleBeforeEach(t)
			rule, _, _ := initSuite(suite)
			suite.mockPaymentSpendRuleRepository.EXPECT().
				Update(gomock.Any(), ruleNamed{name: "Test token", maxAmountPerPayment: "10000"}).
				Return(rule, nil).
				Times(1)
			suite.mockStrategy.EXPECT().
				SaveConfig(gomock.Any(), rule.PolicyId, rule.Id, gomock.Any()).
				Return(nil).
				Times(1)
			suite.mockPaymentSpendRuleWindowRepository.EXPECT().
				Update(gomock.Any(), gomock.Any()).
				Return(&models.PaymentSpendRuleWindow{}, nil).
				Times(1)
			suite.mockPaymentSpendRuleWindowRepository.EXPECT().
				Create(gomock.Any(), gomock.Any()).
				Return(&models.PaymentSpendRuleWindow{}, nil).
				Times(1)
			suite.mockPaymentSpendRuleWindowRepository.EXPECT().
				DeleteById(gomock.Any(), gomock.Any()).
				Return(nil).
				Times(1)

			_, err := suite.sut.Execute(ctx, generateWindowedUpdateRequest(t, rule.Id))

			assert.NoError(t, err)
		})

		t.Run("Should write the rule, its token and its windows inside the unit of work", func(t *testing.T) {
			t.Parallel()

			suite := WhenUpdatingPaymentSpendRuleBeforeEach(t)
			rule, daily, hourly := initSuite(suite)
			suite.mockPaymentSpendRuleRepository.EXPECT().
				Update(suite.unitOfWorkCtx, gomock.Any()).
				Return(rule, nil).
				Times(1)
			suite.mockStrategy.EXPECT().
				SaveConfig(suite.unitOfWorkCtx, rule.PolicyId, rule.Id, gomock.Any()).
				Return(nil).
				Times(1)
			suite.mockPaymentSpendRuleWindowRepository.EXPECT().
				Update(suite.unitOfWorkCtx, gomock.Any()).
				Return(daily, nil).
				Times(1)
			suite.mockPaymentSpendRuleWindowRepository.EXPECT().
				Create(suite.unitOfWorkCtx, gomock.Any()).
				Return(&models.PaymentSpendRuleWindow{}, nil).
				Times(1)
			suite.mockPaymentSpendRuleWindowRepository.EXPECT().
				DeleteById(suite.unitOfWorkCtx, hourly.Id).
				Return(nil).
				Times(1)

			_, err := suite.sut.Execute(ctx, generateWindowedUpdateRequest(t, rule.Id))

			assert.NoError(t, err)
		})

		t.Run("Should keep the spend already recorded against a kept period", func(t *testing.T) {
			t.Parallel()

			suite := WhenUpdatingPaymentSpendRuleBeforeEach(t)
			rule, daily, _ := initSuite(suite)
			suite.mockPaymentSpendRuleRepository.EXPECT().Update(gomock.Any(), gomock.Any()).Return(rule, nil).Times(1)
			suite.mockStrategy.EXPECT().
				SaveConfig(gomock.Any(), rule.PolicyId, rule.Id, gomock.Any()).
				Return(nil).
				Times(1)
			suite.mockPaymentSpendRuleWindowRepository.EXPECT().
				Update(gomock.Any(), windowState{id: daily.Id, maxAmount: "60000", spent: "4000", started: true}).
				Return(daily, nil).
				Times(1)
			suite.mockPaymentSpendRuleWindowRepository.EXPECT().
				Create(gomock.Any(), gomock.Any()).
				Return(&models.PaymentSpendRuleWindow{}, nil).
				Times(1)
			suite.mockPaymentSpendRuleWindowRepository.EXPECT().
				DeleteById(gomock.Any(), gomock.Any()).
				Return(nil).
				Times(1)

			_, err := suite.sut.Execute(ctx, generateWindowedUpdateRequest(t, rule.Id))

			assert.NoError(t, err)
		})

		t.Run("Should create a window for a new period with nothing spent", func(t *testing.T) {
			t.Parallel()

			suite := WhenUpdatingPaymentSpendRuleBeforeEach(t)
			rule, _, _ := initSuite(suite)
			suite.mockPaymentSpendRuleRepository.EXPECT().Update(gomock.Any(), gomock.Any()).Return(rule, nil).Times(1)
			suite.mockStrategy.EXPECT().
				SaveConfig(gomock.Any(), rule.PolicyId, rule.Id, gomock.Any()).
				Return(nil).
				Times(1)
			suite.mockPaymentSpendRuleWindowRepository.EXPECT().
				Update(gomock.Any(), gomock.Any()).
				Return(&models.PaymentSpendRuleWindow{}, nil).
				Times(1)
			suite.mockPaymentSpendRuleWindowRepository.EXPECT().
				Create(gomock.Any(), &models.PaymentSpendRuleWindow{
					PaymentSpendRuleId: rule.Id,
					WindowSeconds:      604800,
					MaxAmount:          "200000",
					Spent:              "0",
				}).
				Return(&models.PaymentSpendRuleWindow{}, nil).
				Times(1)
			suite.mockPaymentSpendRuleWindowRepository.EXPECT().
				DeleteById(gomock.Any(), gomock.Any()).
				Return(nil).
				Times(1)

			_, err := suite.sut.Execute(ctx, generateWindowedUpdateRequest(t, rule.Id))

			assert.NoError(t, err)
		})

		t.Run("Should remove a window no longer requested", func(t *testing.T) {
			t.Parallel()

			suite := WhenUpdatingPaymentSpendRuleBeforeEach(t)
			rule, _, hourly := initSuite(suite)
			suite.mockPaymentSpendRuleRepository.EXPECT().Update(gomock.Any(), gomock.Any()).Return(rule, nil).Times(1)
			suite.mockStrategy.EXPECT().
				SaveConfig(gomock.Any(), rule.PolicyId, rule.Id, gomock.Any()).
				Return(nil).
				Times(1)
			suite.mockPaymentSpendRuleWindowRepository.EXPECT().
				Update(gomock.Any(), gomock.Any()).
				Return(&models.PaymentSpendRuleWindow{}, nil).
				Times(1)
			suite.mockPaymentSpendRuleWindowRepository.EXPECT().
				Create(gomock.Any(), gomock.Any()).
				Return(&models.PaymentSpendRuleWindow{}, nil).
				Times(1)
			suite.mockPaymentSpendRuleWindowRepository.EXPECT().DeleteById(gomock.Any(), hourly.Id).Return(nil).Times(1)

			_, err := suite.sut.Execute(ctx, generateWindowedUpdateRequest(t, rule.Id))

			assert.NoError(t, err)
		})

		t.Run("Should return the rule with its windows", func(t *testing.T) {
			t.Parallel()

			suite := WhenUpdatingPaymentSpendRuleBeforeEach(t)
			rule, daily, _ := initSuite(suite)
			suite.mockPaymentSpendRuleRepository.EXPECT().Update(gomock.Any(), gomock.Any()).Return(rule, nil).Times(1)
			suite.mockStrategy.EXPECT().
				SaveConfig(gomock.Any(), rule.PolicyId, rule.Id, gomock.Any()).
				Return(nil).
				Times(1)
			suite.mockPaymentSpendRuleWindowRepository.EXPECT().
				Update(gomock.Any(), gomock.Any()).
				Return(daily, nil).
				Times(1)
			suite.mockPaymentSpendRuleWindowRepository.EXPECT().
				Create(gomock.Any(), gomock.Any()).
				Return(model_builders.NewPaymentSpendRuleWindowBuilder().
					WithPaymentSpendRuleId(rule.Id).
					WithWindowSeconds(604800).
					WithMaxAmount("200000").
					Build(), nil).
				Times(1)
			suite.mockPaymentSpendRuleWindowRepository.EXPECT().
				DeleteById(gomock.Any(), gomock.Any()).
				Return(nil).
				Times(1)

			response, err := suite.sut.Execute(ctx, generateWindowedUpdateRequest(t, rule.Id))

			assert.NoError(t, err)
			assert.Equal(t, rule.Id, response.Id)
			if assert.Len(t, response.Windows, 2) {
				assert.Equal(t, "50000", response.Windows[0].MaxAmount)
				assert.Equal(t, "200000", response.Windows[1].MaxAmount)
			}
		})
	})

	t.Run("Given the rule moves to another token", func(t *testing.T) {
		t.Parallel()

		initSuite := func(
			suite *WhenUpdatingPaymentSpendRuleTestingSuite,
		) (*models.PaymentSpendRule, *models.PaymentSpendRuleWindow) {
			rule := model_builders.NewPaymentSpendRuleBuilder().Build()
			daily := model_builders.NewPaymentSpendRuleWindowBuilder().
				WithPaymentSpendRuleId(rule.Id).
				WithMaxAmount("50000").
				WithSpent("4000").
				WithStartedAt(time.Now().Add(-time.Hour)).
				Build()

			suite.mockStrategy.EXPECT().Validate(gomock.Any(), gomock.Any()).Return(nil).Times(1)
			suite.mockPaymentSpendRuleRepository.EXPECT().GetById(gomock.Any(), rule.Id).Return(rule, nil).Times(1)
			suite.mockStrategy.EXPECT().
				RuleIdForAsset(gomock.Any(), rule.PolicyId, gomock.Any()).
				Return(uuid.Nil, nil).
				Times(1)
			suite.mockUnitOfWorkFactory.EXPECT().Start(gomock.Any()).Return(suite.mockUnitOfWork, nil).Times(1)
			suite.mockPaymentSpendRuleRepository.EXPECT().Update(gomock.Any(), gomock.Any()).Return(rule, nil).Times(1)
			suite.mockStrategy.EXPECT().
				SaveConfig(gomock.Any(), rule.PolicyId, rule.Id, gomock.Any()).
				Return(nil).
				Times(1)
			suite.mockPaymentSpendRuleWindowRepository.EXPECT().
				ListByRuleIdForUpdate(gomock.Any(), rule.Id).
				Return([]models.PaymentSpendRuleWindow{*daily}, nil).
				Times(1)
			suite.mockUnitOfWork.EXPECT().Commit().Return(nil).Times(1)
			return rule, daily
		}

		// Spend recorded in the old token says nothing about the new one.
		t.Run("Should restart every kept period", func(t *testing.T) {
			t.Parallel()

			suite := WhenUpdatingPaymentSpendRuleBeforeEach(t)
			rule, daily := initSuite(suite)
			suite.mockPaymentSpendRuleWindowRepository.EXPECT().
				Update(gomock.Any(), windowState{id: daily.Id, maxAmount: "50000", spent: "0", started: false}).
				Return(daily, nil).
				Times(1)

			_, err := suite.sut.Execute(ctx, generateUpdatePaymentSpendRuleRequest(t, rule.Id))

			assert.NoError(t, err)
		})
	})

	t.Run("Given the update drops the per-payment cap", func(t *testing.T) {
		t.Parallel()

		initSuite := func(suite *WhenUpdatingPaymentSpendRuleTestingSuite) *models.PaymentSpendRule {
			rule := model_builders.NewPaymentSpendRuleBuilder().WithMaxAmountPerPayment("10000").Build()

			suite.mockStrategy.EXPECT().Validate(gomock.Any(), gomock.Any()).Return(nil).Times(1)
			suite.mockPaymentSpendRuleRepository.EXPECT().GetById(gomock.Any(), rule.Id).Return(rule, nil).Times(1)
			suite.mockStrategy.EXPECT().
				RuleIdForAsset(gomock.Any(), rule.PolicyId, gomock.Any()).
				Return(rule.Id, nil).
				Times(1)
			suite.mockUnitOfWorkFactory.EXPECT().Start(gomock.Any()).Return(suite.mockUnitOfWork, nil).Times(1)
			suite.mockStrategy.EXPECT().
				SaveConfig(gomock.Any(), rule.PolicyId, rule.Id, gomock.Any()).
				Return(nil).
				Times(1)
			suite.mockPaymentSpendRuleWindowRepository.EXPECT().
				ListByRuleIdForUpdate(gomock.Any(), rule.Id).
				Return([]models.PaymentSpendRuleWindow{}, nil).
				Times(1)
			suite.mockPaymentSpendRuleWindowRepository.EXPECT().
				Create(gomock.Any(), gomock.Any()).
				Return(&models.PaymentSpendRuleWindow{}, nil).
				Times(1)
			suite.mockUnitOfWork.EXPECT().Commit().Return(nil).Times(1)
			return rule
		}

		t.Run("Should clear the cap", func(t *testing.T) {
			t.Parallel()

			suite := WhenUpdatingPaymentSpendRuleBeforeEach(t)
			rule := initSuite(suite)
			suite.mockPaymentSpendRuleRepository.EXPECT().
				Update(gomock.Any(), ruleUncapped{}).
				Return(rule, nil).
				Times(1)
			request := generateUpdatePaymentSpendRuleRequest(t, rule.Id)
			request.MaxAmountPerPayment = nil

			_, err := suite.sut.Execute(ctx, request)

			assert.NoError(t, err)
		})
	})
}

type ruleNamed struct {
	name                string
	maxAmountPerPayment string
}

func (m ruleNamed) Matches(x any) bool {
	rule, ok := x.(*models.PaymentSpendRule)
	return ok && rule.Name == m.name &&
		rule.MaxAmountPerPayment != nil && *rule.MaxAmountPerPayment == m.maxAmountPerPayment
}

func (m ruleNamed) String() string {
	return "rule named " + m.name + " capped at " + m.maxAmountPerPayment
}

type windowState struct {
	id        uuid.UUID
	maxAmount string
	spent     string
	started   bool
}

func (m windowState) Matches(x any) bool {
	window, ok := x.(*models.PaymentSpendRuleWindow)
	return ok && window.Id == m.id && window.MaxAmount == m.maxAmount && window.Spent == m.spent &&
		(window.StartedAt != nil) == m.started
}

func (m windowState) String() string {
	return "window " + m.id.String() + " at " + m.maxAmount + " with spent " + m.spent
}

type ruleUncapped struct{}

func (m ruleUncapped) Matches(x any) bool {
	rule, ok := x.(*models.PaymentSpendRule)
	return ok && rule.MaxAmountPerPayment == nil
}

func (m ruleUncapped) String() string {
	return "rule without a per-payment cap"
}
