package use_cases_test

import (
	"context"
	"testing"

	"github.com/6022-labs/agentic-agent-node/src/agent_payment/models"
	"github.com/6022-labs/agentic-agent-node/src/agent_payment/use_cases"
	"github.com/6022-labs/agentic-agent-node/src/agent_payment/use_cases/requests"
	apperrors "github.com/6022-labs/agentic-agent-node/src/common/errors"
	payment_repositories_mocks "github.com/6022-labs/agentic-agent-node/tests/agent_payment_mocks/repositories_mocks"
	payment_interfaces_mocks "github.com/6022-labs/agentic-agent-node/tests/agent_payment_mocks/services_mocks/interfaces_mocks"
	"github.com/6022-labs/agentic-agent-node/tests/common_tests/model_builders"
	runtime_common_mocks "github.com/6022-labs/agentic-agent-node/tests/runtime_common_mocks/services_mocks/interfaces_mocks"
	"github.com/google/uuid"
	"github.com/samber/lo"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

type WhenUpdatingPaymentReceiveRuleTestingSuite struct {
	sut *use_cases.UpdatePaymentReceiveRule

	mockPaymentReceivePolicyProvider      *payment_interfaces_mocks.MockPaymentReceivePolicyProviderInterface
	mockPaymentReceiveRuleRepository      *payment_repositories_mocks.MockPaymentReceiveRuleRepositoryInterface
	mockPaymentReceiveRulePriceRepository *payment_repositories_mocks.MockPaymentReceiveRulePriceRepositoryInterface
	mockStrategy                          *payment_interfaces_mocks.MockPaymentReceiveMethodStrategyInterface
	mockUnitOfWorkFactory                 *runtime_common_mocks.MockUnitOfWorkFactoryInterface
	mockUnitOfWork                        *runtime_common_mocks.MockUnitOfWorkInterface
	unitOfWorkCtx                         context.Context
}

func WhenUpdatingPaymentReceiveRuleBeforeEach(t *testing.T) *WhenUpdatingPaymentReceiveRuleTestingSuite {
	mockController := gomock.NewController(t)

	mockPaymentReceivePolicyProvider := payment_interfaces_mocks.NewMockPaymentReceivePolicyProviderInterface(
		mockController,
	)
	mockPaymentReceiveRuleRepository := payment_repositories_mocks.NewMockPaymentReceiveRuleRepositoryInterface(
		mockController,
	)
	mockPaymentReceiveRulePriceRepository :=
		payment_repositories_mocks.NewMockPaymentReceiveRulePriceRepositoryInterface(mockController)
	strategies, mockStrategy := newMockStrategies(mockController)
	mockUnitOfWorkFactory := runtime_common_mocks.NewMockUnitOfWorkFactoryInterface(mockController)
	mockUnitOfWork, unitOfWorkCtx := newMockUnitOfWork(mockController)

	return &WhenUpdatingPaymentReceiveRuleTestingSuite{
		sut: use_cases.NewUpdatePaymentReceiveRule(
			mockPaymentReceivePolicyProvider,
			mockPaymentReceiveRuleRepository,
			mockPaymentReceiveRulePriceRepository,
			strategies,
			mockUnitOfWorkFactory,
		),

		mockPaymentReceivePolicyProvider:      mockPaymentReceivePolicyProvider,
		mockPaymentReceiveRuleRepository:      mockPaymentReceiveRuleRepository,
		mockPaymentReceiveRulePriceRepository: mockPaymentReceiveRulePriceRepository,
		mockStrategy:                          mockStrategy,
		mockUnitOfWorkFactory:                 mockUnitOfWorkFactory,
		mockUnitOfWork:                        mockUnitOfWork,
		unitOfWorkCtx:                         unitOfWorkCtx,
	}
}

func generateUpdatePaymentReceiveRuleRequest(id uuid.UUID) *requests.UpdatePaymentReceiveRuleRequest {
	return &requests.UpdatePaymentReceiveRuleRequest{
		Id:                            id,
		SavePaymentReceiveRuleRequest: *generateSavePaymentReceiveRuleRequest(),
	}
}

func TestWhenUpdatingPaymentReceiveRule(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	t.Run("Given the request is invalid", func(t *testing.T) {
		t.Parallel()

		t.Run("Should return a validation error before touching the rule", func(t *testing.T) {
			t.Parallel()

			suite := WhenUpdatingPaymentReceiveRuleBeforeEach(t)
			request := generateUpdatePaymentReceiveRuleRequest(uuid.New())
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

			suite := WhenUpdatingPaymentReceiveRuleBeforeEach(t)
			sut := use_cases.NewUpdatePaymentReceiveRule(
				suite.mockPaymentReceivePolicyProvider,
				suite.mockPaymentReceiveRuleRepository,
				suite.mockPaymentReceiveRulePriceRepository,
				newUnsupportedMethodStrategies(gomock.NewController(t)),
				suite.mockUnitOfWorkFactory,
			)

			rule, err := sut.Execute(ctx, generateUpdatePaymentReceiveRuleRequest(uuid.New()))

			assert.Nil(t, rule)
			var validationErr *apperrors.ValidationError
			assert.ErrorAs(t, err, &validationErr)
			assert.Equal(t, "method", validationErr.Field)
		})
	})

	t.Run("Given the strategy rejects the config", func(t *testing.T) {
		t.Parallel()

		initSuite := func(suite *WhenUpdatingPaymentReceiveRuleTestingSuite) {
			suite.mockStrategy.EXPECT().
				Validate(gomock.Any(), gomock.Any()).
				Return(apperrors.NewValidationError("config.network", "unsupported")).
				Times(1)
		}

		t.Run("Should return a validation error on the method field", func(t *testing.T) {
			t.Parallel()

			suite := WhenUpdatingPaymentReceiveRuleBeforeEach(t)
			initSuite(suite)

			rule, err := suite.sut.Execute(ctx, generateUpdatePaymentReceiveRuleRequest(uuid.New()))

			assert.Nil(t, rule)
			var validationErr *apperrors.ValidationError
			assert.ErrorAs(t, err, &validationErr)
			assert.Equal(t, "config.network", validationErr.Field)
		})
	})

	t.Run("Given the rule lookup fails", func(t *testing.T) {
		t.Parallel()

		initSuite := func(suite *WhenUpdatingPaymentReceiveRuleTestingSuite) {
			suite.mockStrategy.EXPECT().Validate(gomock.Any(), gomock.Any()).Return(nil).Times(1)
			suite.mockPaymentReceiveRuleRepository.EXPECT().
				GetById(gomock.Any(), gomock.Any()).
				Return(nil, assert.AnError).
				Times(1)
		}

		t.Run("Should return an InternalError", func(t *testing.T) {
			t.Parallel()

			suite := WhenUpdatingPaymentReceiveRuleBeforeEach(t)
			initSuite(suite)

			rule, err := suite.sut.Execute(ctx, generateUpdatePaymentReceiveRuleRequest(uuid.New()))

			assert.Nil(t, rule)
			var internalErr *apperrors.InternalError
			assert.ErrorAs(t, err, &internalErr)
			assert.Equal(t, "update_failed", internalErr.Code)
		})
	})

	t.Run("Given the rule does not exist", func(t *testing.T) {
		t.Parallel()

		initSuite := func(suite *WhenUpdatingPaymentReceiveRuleTestingSuite) {
			suite.mockStrategy.EXPECT().Validate(gomock.Any(), gomock.Any()).Return(nil).Times(1)
			suite.mockPaymentReceiveRuleRepository.EXPECT().
				GetById(gomock.Any(), gomock.Any()).
				Return(nil, nil).
				Times(1)
		}

		t.Run("Should return a not found error", func(t *testing.T) {
			t.Parallel()

			suite := WhenUpdatingPaymentReceiveRuleBeforeEach(t)
			initSuite(suite)

			rule, err := suite.sut.Execute(ctx, generateUpdatePaymentReceiveRuleRequest(uuid.New()))

			assert.Nil(t, rule)
			var notFoundErr *apperrors.NotFoundError
			assert.ErrorAs(t, err, &notFoundErr)
			assert.Equal(t, "rule_not_found", notFoundErr.Code)
		})
	})

	t.Run("Given looking up the token's rule fails", func(t *testing.T) {
		t.Parallel()

		initSuite := func(suite *WhenUpdatingPaymentReceiveRuleTestingSuite) *models.PaymentReceiveRule {
			rule := model_builders.NewPaymentReceiveRuleBuilder().Build()

			suite.mockStrategy.EXPECT().Validate(gomock.Any(), gomock.Any()).Return(nil).Times(1)
			suite.mockPaymentReceiveRuleRepository.EXPECT().
				GetById(gomock.Any(), rule.Id).
				Return(lo.ToPtr(*rule), nil).
				Times(1)
			suite.mockStrategy.EXPECT().
				RuleIdForAsset(gomock.Any(), rule.PolicyId, gomock.Any()).
				Return(uuid.Nil, assert.AnError).
				Times(1)
			return rule
		}

		t.Run("Should return an InternalError", func(t *testing.T) {
			t.Parallel()

			suite := WhenUpdatingPaymentReceiveRuleBeforeEach(t)
			rule := initSuite(suite)

			updated, err := suite.sut.Execute(ctx, generateUpdatePaymentReceiveRuleRequest(rule.Id))

			assert.Nil(t, updated)
			var internalErr *apperrors.InternalError
			assert.ErrorAs(t, err, &internalErr)
			assert.Equal(t, "rule_lookup_failed", internalErr.Code)
		})
	})

	t.Run("Given another rule already prices the token", func(t *testing.T) {
		t.Parallel()

		initSuite := func(suite *WhenUpdatingPaymentReceiveRuleTestingSuite) *models.PaymentReceiveRule {
			rule := model_builders.NewPaymentReceiveRuleBuilder().Build()

			suite.mockStrategy.EXPECT().Validate(gomock.Any(), gomock.Any()).Return(nil).Times(1)
			suite.mockPaymentReceiveRuleRepository.EXPECT().
				GetById(gomock.Any(), rule.Id).
				Return(lo.ToPtr(*rule), nil).
				Times(1)
			suite.mockStrategy.EXPECT().
				RuleIdForAsset(gomock.Any(), rule.PolicyId, gomock.Any()).
				Return(uuid.New(), nil).
				Times(1)
			return rule
		}

		t.Run("Should return a conflict error", func(t *testing.T) {
			t.Parallel()

			suite := WhenUpdatingPaymentReceiveRuleBeforeEach(t)
			rule := initSuite(suite)

			updated, err := suite.sut.Execute(ctx, generateUpdatePaymentReceiveRuleRequest(rule.Id))

			assert.Nil(t, updated)
			var conflictErr *apperrors.ConflictError
			assert.ErrorAs(t, err, &conflictErr)
			assert.Equal(t, "asset_already_has_rule", conflictErr.Code)
		})
	})

	t.Run("Given starting the unit of work fails", func(t *testing.T) {
		t.Parallel()

		initSuite := func(suite *WhenUpdatingPaymentReceiveRuleTestingSuite) *models.PaymentReceiveRule {
			rule := model_builders.NewPaymentReceiveRuleBuilder().Build()

			suite.mockStrategy.EXPECT().Validate(gomock.Any(), gomock.Any()).Return(nil).Times(1)
			suite.mockPaymentReceiveRuleRepository.EXPECT().
				GetById(gomock.Any(), rule.Id).
				Return(lo.ToPtr(*rule), nil).
				Times(1)
			suite.mockStrategy.EXPECT().
				RuleIdForAsset(gomock.Any(), rule.PolicyId, gomock.Any()).
				Return(rule.Id, nil).
				Times(1)
			suite.mockUnitOfWorkFactory.EXPECT().Start(gomock.Any()).Return(nil, assert.AnError).Times(1)
			return rule
		}

		t.Run("Should return an InternalError", func(t *testing.T) {
			t.Parallel()

			suite := WhenUpdatingPaymentReceiveRuleBeforeEach(t)
			rule := initSuite(suite)

			updated, err := suite.sut.Execute(ctx, generateUpdatePaymentReceiveRuleRequest(rule.Id))

			assert.Nil(t, updated)
			var internalErr *apperrors.InternalError
			assert.ErrorAs(t, err, &internalErr)
			assert.Equal(t, "update_failed", internalErr.Code)
		})
	})

	t.Run("Given the repository fails to persist the rule", func(t *testing.T) {
		t.Parallel()

		initSuite := func(suite *WhenUpdatingPaymentReceiveRuleTestingSuite) *models.PaymentReceiveRule {
			rule := model_builders.NewPaymentReceiveRuleBuilder().Build()

			suite.mockStrategy.EXPECT().Validate(gomock.Any(), gomock.Any()).Return(nil).Times(1)
			suite.mockPaymentReceiveRuleRepository.EXPECT().
				GetById(gomock.Any(), rule.Id).
				Return(lo.ToPtr(*rule), nil).
				Times(1)
			suite.mockStrategy.EXPECT().
				RuleIdForAsset(gomock.Any(), rule.PolicyId, gomock.Any()).
				Return(rule.Id, nil).
				Times(1)
			suite.mockUnitOfWorkFactory.EXPECT().Start(gomock.Any()).Return(suite.mockUnitOfWork, nil).Times(1)
			suite.mockPaymentReceiveRuleRepository.EXPECT().
				Update(gomock.Any(), gomock.Any()).
				Return(nil, assert.AnError).
				Times(1)
			return rule
		}

		t.Run("Should return an InternalError", func(t *testing.T) {
			t.Parallel()

			suite := WhenUpdatingPaymentReceiveRuleBeforeEach(t)
			rule := initSuite(suite)

			updated, err := suite.sut.Execute(ctx, generateUpdatePaymentReceiveRuleRequest(rule.Id))

			assert.Nil(t, updated)
			var internalErr *apperrors.InternalError
			assert.ErrorAs(t, err, &internalErr)
			assert.Equal(t, "update_failed", internalErr.Code)
		})
	})

	t.Run("Given persisting the config fails", func(t *testing.T) {
		t.Parallel()

		initSuite := func(suite *WhenUpdatingPaymentReceiveRuleTestingSuite) *models.PaymentReceiveRule {
			rule := model_builders.NewPaymentReceiveRuleBuilder().Build()

			suite.mockStrategy.EXPECT().Validate(gomock.Any(), gomock.Any()).Return(nil).Times(1)
			suite.mockPaymentReceiveRuleRepository.EXPECT().
				GetById(gomock.Any(), rule.Id).
				Return(lo.ToPtr(*rule), nil).
				Times(1)
			suite.mockStrategy.EXPECT().
				RuleIdForAsset(gomock.Any(), rule.PolicyId, gomock.Any()).
				Return(rule.Id, nil).
				Times(1)
			suite.mockUnitOfWorkFactory.EXPECT().Start(gomock.Any()).Return(suite.mockUnitOfWork, nil).Times(1)
			suite.mockPaymentReceiveRuleRepository.EXPECT().
				Update(gomock.Any(), gomock.Any()).
				Return(rule, nil).
				Times(1)
			suite.mockStrategy.EXPECT().
				SaveConfig(gomock.Any(), rule.PolicyId, rule.Id, gomock.Any()).
				Return(assert.AnError).
				Times(1)
			return rule
		}

		t.Run("Should return an InternalError", func(t *testing.T) {
			t.Parallel()

			suite := WhenUpdatingPaymentReceiveRuleBeforeEach(t)
			rule := initSuite(suite)

			updated, err := suite.sut.Execute(ctx, generateUpdatePaymentReceiveRuleRequest(rule.Id))

			assert.Nil(t, updated)
			var internalErr *apperrors.InternalError
			assert.ErrorAs(t, err, &internalErr)
			assert.Equal(t, "update_failed", internalErr.Code)
		})
	})

	t.Run("Given clearing the old prices fails", func(t *testing.T) {
		t.Parallel()

		initSuite := func(suite *WhenUpdatingPaymentReceiveRuleTestingSuite) *models.PaymentReceiveRule {
			rule := model_builders.NewPaymentReceiveRuleBuilder().Build()

			suite.mockStrategy.EXPECT().Validate(gomock.Any(), gomock.Any()).Return(nil).Times(1)
			suite.mockPaymentReceiveRuleRepository.EXPECT().
				GetById(gomock.Any(), rule.Id).
				Return(lo.ToPtr(*rule), nil).
				Times(1)
			suite.mockStrategy.EXPECT().
				RuleIdForAsset(gomock.Any(), rule.PolicyId, gomock.Any()).
				Return(rule.Id, nil).
				Times(1)
			suite.mockUnitOfWorkFactory.EXPECT().Start(gomock.Any()).Return(suite.mockUnitOfWork, nil).Times(1)
			suite.mockPaymentReceiveRuleRepository.EXPECT().
				Update(gomock.Any(), gomock.Any()).
				Return(rule, nil).
				Times(1)
			suite.mockStrategy.EXPECT().
				SaveConfig(gomock.Any(), rule.PolicyId, rule.Id, gomock.Any()).
				Return(nil).
				Times(1)
			suite.mockPaymentReceiveRulePriceRepository.EXPECT().
				DeleteByRuleId(gomock.Any(), rule.Id).
				Return(assert.AnError).
				Times(1)
			return rule
		}

		t.Run("Should return an InternalError", func(t *testing.T) {
			t.Parallel()

			suite := WhenUpdatingPaymentReceiveRuleBeforeEach(t)
			rule := initSuite(suite)

			updated, err := suite.sut.Execute(ctx, generateUpdatePaymentReceiveRuleRequest(rule.Id))

			assert.Nil(t, updated)
			var internalErr *apperrors.InternalError
			assert.ErrorAs(t, err, &internalErr)
			assert.Equal(t, "update_failed", internalErr.Code)
		})
	})

	t.Run("Given persisting a price fails", func(t *testing.T) {
		t.Parallel()

		initSuite := func(suite *WhenUpdatingPaymentReceiveRuleTestingSuite) *models.PaymentReceiveRule {
			rule := model_builders.NewPaymentReceiveRuleBuilder().Build()

			suite.mockStrategy.EXPECT().Validate(gomock.Any(), gomock.Any()).Return(nil).Times(1)
			suite.mockPaymentReceiveRuleRepository.EXPECT().
				GetById(gomock.Any(), rule.Id).
				Return(lo.ToPtr(*rule), nil).
				Times(1)
			suite.mockStrategy.EXPECT().
				RuleIdForAsset(gomock.Any(), rule.PolicyId, gomock.Any()).
				Return(rule.Id, nil).
				Times(1)
			suite.mockUnitOfWorkFactory.EXPECT().Start(gomock.Any()).Return(suite.mockUnitOfWork, nil).Times(1)
			suite.mockPaymentReceiveRuleRepository.EXPECT().
				Update(gomock.Any(), gomock.Any()).
				Return(rule, nil).
				Times(1)
			suite.mockStrategy.EXPECT().
				SaveConfig(gomock.Any(), rule.PolicyId, rule.Id, gomock.Any()).
				Return(nil).
				Times(1)
			suite.mockPaymentReceiveRulePriceRepository.EXPECT().
				DeleteByRuleId(gomock.Any(), rule.Id).
				Return(nil).
				Times(1)
			suite.mockPaymentReceiveRulePriceRepository.EXPECT().
				Create(gomock.Any(), gomock.Any()).
				Return(nil, assert.AnError).
				Times(1)
			return rule
		}

		t.Run("Should return an InternalError", func(t *testing.T) {
			t.Parallel()

			suite := WhenUpdatingPaymentReceiveRuleBeforeEach(t)
			rule := initSuite(suite)

			updated, err := suite.sut.Execute(ctx, generateUpdatePaymentReceiveRuleRequest(rule.Id))

			assert.Nil(t, updated)
			var internalErr *apperrors.InternalError
			assert.ErrorAs(t, err, &internalErr)
			assert.Equal(t, "update_failed", internalErr.Code)
		})
	})

	t.Run("Given committing the rule fails", func(t *testing.T) {
		t.Parallel()

		initSuite := func(suite *WhenUpdatingPaymentReceiveRuleTestingSuite) *models.PaymentReceiveRule {
			rule := model_builders.NewPaymentReceiveRuleBuilder().Build()
			price := model_builders.NewPaymentReceiveRulePriceBuilder().WithPaymentReceiveRuleId(rule.Id).Build()

			suite.mockStrategy.EXPECT().Validate(gomock.Any(), gomock.Any()).Return(nil).Times(1)
			suite.mockPaymentReceiveRuleRepository.EXPECT().
				GetById(gomock.Any(), rule.Id).
				Return(lo.ToPtr(*rule), nil).
				Times(1)
			suite.mockStrategy.EXPECT().
				RuleIdForAsset(gomock.Any(), rule.PolicyId, gomock.Any()).
				Return(rule.Id, nil).
				Times(1)
			suite.mockUnitOfWorkFactory.EXPECT().Start(gomock.Any()).Return(suite.mockUnitOfWork, nil).Times(1)
			suite.mockPaymentReceiveRuleRepository.EXPECT().
				Update(gomock.Any(), gomock.Any()).
				Return(rule, nil).
				Times(1)
			suite.mockStrategy.EXPECT().
				SaveConfig(gomock.Any(), rule.PolicyId, rule.Id, gomock.Any()).
				Return(nil).
				Times(1)
			suite.mockPaymentReceiveRulePriceRepository.EXPECT().
				DeleteByRuleId(gomock.Any(), rule.Id).
				Return(nil).
				Times(1)
			suite.mockPaymentReceiveRulePriceRepository.EXPECT().
				Create(gomock.Any(), gomock.Any()).
				Return(price, nil).
				Times(1)
			suite.mockUnitOfWork.EXPECT().Commit().Return(assert.AnError).Times(1)
			return rule
		}

		t.Run("Should return an InternalError", func(t *testing.T) {
			t.Parallel()

			suite := WhenUpdatingPaymentReceiveRuleBeforeEach(t)
			rule := initSuite(suite)

			updated, err := suite.sut.Execute(ctx, generateUpdatePaymentReceiveRuleRequest(rule.Id))

			assert.Nil(t, updated)
			var internalErr *apperrors.InternalError
			assert.ErrorAs(t, err, &internalErr)
			assert.Equal(t, "update_failed", internalErr.Code)
		})
	})

	t.Run("Given the rule keeps its token", func(t *testing.T) {
		t.Parallel()

		rule := model_builders.NewPaymentReceiveRuleBuilder().WithName("Old name").Build()
		price := model_builders.NewPaymentReceiveRulePriceBuilder().
			WithPaymentReceiveRuleId(rule.Id).
			WithPrice("10000").
			WithAddressPatterns([]string{".*"}).
			Build()

		initSuite := func(suite *WhenUpdatingPaymentReceiveRuleTestingSuite) {
			suite.mockStrategy.EXPECT().Validate(ctx, gomock.Any()).Return(nil).Times(1)
			suite.mockPaymentReceiveRuleRepository.EXPECT().GetById(ctx, rule.Id).Return(lo.ToPtr(*rule), nil).Times(1)
			suite.mockStrategy.EXPECT().RuleIdForAsset(ctx, rule.PolicyId, gomock.Any()).Return(rule.Id, nil).Times(1)
			suite.mockUnitOfWorkFactory.EXPECT().Start(ctx).Return(suite.mockUnitOfWork, nil).Times(1)
		}

		t.Run("Should replace the rule's fields", func(t *testing.T) {
			t.Parallel()

			suite := WhenUpdatingPaymentReceiveRuleBeforeEach(t)
			initSuite(suite)
			suite.mockPaymentReceiveRuleRepository.EXPECT().
				Update(gomock.Any(), gomock.Cond(func(updated *models.PaymentReceiveRule) bool {
					return updated.Id == rule.Id && updated.Name == "Public price"
				})).
				Return(rule, nil).
				Times(1)
			suite.mockStrategy.EXPECT().
				SaveConfig(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
				Return(nil).
				Times(1)
			suite.mockPaymentReceiveRulePriceRepository.EXPECT().
				DeleteByRuleId(gomock.Any(), rule.Id).
				Return(nil).
				Times(1)
			suite.mockPaymentReceiveRulePriceRepository.EXPECT().
				Create(gomock.Any(), gomock.Any()).
				Return(price, nil).
				Times(1)
			suite.mockUnitOfWork.EXPECT().Commit().Return(nil).Times(1)

			_, err := suite.sut.Execute(ctx, generateUpdatePaymentReceiveRuleRequest(rule.Id))

			assert.NoError(t, err)
		})

		t.Run("Should persist the config for the rule", func(t *testing.T) {
			t.Parallel()

			suite := WhenUpdatingPaymentReceiveRuleBeforeEach(t)
			initSuite(suite)
			suite.mockPaymentReceiveRuleRepository.EXPECT().
				Update(gomock.Any(), gomock.Any()).
				Return(rule, nil).
				Times(1)
			suite.mockStrategy.EXPECT().
				SaveConfig(gomock.Any(), rule.PolicyId, rule.Id, gomock.Any()).
				Return(nil).
				Times(1)
			suite.mockPaymentReceiveRulePriceRepository.EXPECT().
				DeleteByRuleId(gomock.Any(), rule.Id).
				Return(nil).
				Times(1)
			suite.mockPaymentReceiveRulePriceRepository.EXPECT().
				Create(gomock.Any(), gomock.Any()).
				Return(price, nil).
				Times(1)
			suite.mockUnitOfWork.EXPECT().Commit().Return(nil).Times(1)

			_, err := suite.sut.Execute(ctx, generateUpdatePaymentReceiveRuleRequest(rule.Id))

			assert.NoError(t, err)
		})

		t.Run("Should replace the stored prices with the requested ones", func(t *testing.T) {
			t.Parallel()

			suite := WhenUpdatingPaymentReceiveRuleBeforeEach(t)
			initSuite(suite)
			suite.mockPaymentReceiveRuleRepository.EXPECT().
				Update(gomock.Any(), gomock.Any()).
				Return(rule, nil).
				Times(1)
			suite.mockStrategy.EXPECT().
				SaveConfig(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
				Return(nil).
				Times(1)
			gomock.InOrder(
				suite.mockPaymentReceiveRulePriceRepository.EXPECT().
					DeleteByRuleId(gomock.Any(), rule.Id).
					Return(nil).
					Times(1),
				suite.mockPaymentReceiveRulePriceRepository.EXPECT().
					Create(gomock.Any(), &models.PaymentReceiveRulePrice{
						PaymentReceiveRuleId: rule.Id,
						Price:                "10000",
						AddressPatterns:      []string{".*"},
					}).
					Return(price, nil).
					Times(1),
			)
			suite.mockUnitOfWork.EXPECT().Commit().Return(nil).Times(1)

			_, err := suite.sut.Execute(ctx, generateUpdatePaymentReceiveRuleRequest(rule.Id))

			assert.NoError(t, err)
		})

		t.Run("Should save the rule, its config and its prices in one transaction", func(t *testing.T) {
			t.Parallel()

			suite := WhenUpdatingPaymentReceiveRuleBeforeEach(t)
			initSuite(suite)
			suite.mockPaymentReceiveRuleRepository.EXPECT().
				Update(suite.unitOfWorkCtx, gomock.Any()).
				Return(rule, nil).
				Times(1)
			suite.mockStrategy.EXPECT().
				SaveConfig(suite.unitOfWorkCtx, rule.PolicyId, rule.Id, gomock.Any()).
				Return(nil).
				Times(1)
			suite.mockPaymentReceiveRulePriceRepository.EXPECT().
				DeleteByRuleId(suite.unitOfWorkCtx, rule.Id).
				Return(nil).
				Times(1)
			suite.mockPaymentReceiveRulePriceRepository.EXPECT().
				Create(suite.unitOfWorkCtx, gomock.Any()).
				Return(price, nil).
				Times(1)
			suite.mockUnitOfWork.EXPECT().Commit().Return(nil).Times(1)

			_, err := suite.sut.Execute(ctx, generateUpdatePaymentReceiveRuleRequest(rule.Id))

			assert.NoError(t, err)
		})

		t.Run("Should return the rule with its prices", func(t *testing.T) {
			t.Parallel()

			suite := WhenUpdatingPaymentReceiveRuleBeforeEach(t)
			initSuite(suite)
			suite.mockPaymentReceiveRuleRepository.EXPECT().
				Update(gomock.Any(), gomock.Any()).
				Return(rule, nil).
				Times(1)
			suite.mockStrategy.EXPECT().
				SaveConfig(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
				Return(nil).
				Times(1)
			suite.mockPaymentReceiveRulePriceRepository.EXPECT().
				DeleteByRuleId(gomock.Any(), rule.Id).
				Return(nil).
				Times(1)
			suite.mockPaymentReceiveRulePriceRepository.EXPECT().
				Create(gomock.Any(), gomock.Any()).
				Return(price, nil).
				Times(1)
			suite.mockUnitOfWork.EXPECT().Commit().Return(nil).Times(1)

			response, err := suite.sut.Execute(ctx, generateUpdatePaymentReceiveRuleRequest(rule.Id))

			assert.NoError(t, err)
			assert.Equal(t, rule.Id, response.Id)
			if assert.Len(t, response.Prices, 1) {
				assert.Equal(t, "10000", response.Prices[0].Price)
			}
		})
	})

	t.Run("Given the rule moves to a token no other rule prices", func(t *testing.T) {
		t.Parallel()

		rule := model_builders.NewPaymentReceiveRuleBuilder().Build()
		price := model_builders.NewPaymentReceiveRulePriceBuilder().WithPaymentReceiveRuleId(rule.Id).Build()

		initSuite := func(suite *WhenUpdatingPaymentReceiveRuleTestingSuite) {
			suite.mockStrategy.EXPECT().Validate(gomock.Any(), gomock.Any()).Return(nil).Times(1)
			suite.mockPaymentReceiveRuleRepository.EXPECT().
				GetById(gomock.Any(), rule.Id).
				Return(lo.ToPtr(*rule), nil).
				Times(1)
			suite.mockStrategy.EXPECT().
				RuleIdForAsset(gomock.Any(), rule.PolicyId, gomock.Any()).
				Return(uuid.Nil, nil).
				Times(1)
			suite.mockUnitOfWorkFactory.EXPECT().Start(gomock.Any()).Return(suite.mockUnitOfWork, nil).Times(1)
			suite.mockPaymentReceiveRuleRepository.EXPECT().
				Update(gomock.Any(), gomock.Any()).
				Return(rule, nil).
				Times(1)
			suite.mockStrategy.EXPECT().
				SaveConfig(gomock.Any(), rule.PolicyId, rule.Id, gomock.Any()).
				Return(nil).
				Times(1)
			suite.mockPaymentReceiveRulePriceRepository.EXPECT().
				DeleteByRuleId(gomock.Any(), rule.Id).
				Return(nil).
				Times(1)
			suite.mockPaymentReceiveRulePriceRepository.EXPECT().
				Create(gomock.Any(), gomock.Any()).
				Return(price, nil).
				Times(1)
			suite.mockUnitOfWork.EXPECT().Commit().Return(nil).Times(1)
		}

		t.Run("Should update the rule", func(t *testing.T) {
			t.Parallel()

			suite := WhenUpdatingPaymentReceiveRuleBeforeEach(t)
			initSuite(suite)

			response, err := suite.sut.Execute(ctx, generateUpdatePaymentReceiveRuleRequest(rule.Id))

			assert.NoError(t, err)
			assert.Equal(t, rule.Id, response.Id)
		})
	})
}
