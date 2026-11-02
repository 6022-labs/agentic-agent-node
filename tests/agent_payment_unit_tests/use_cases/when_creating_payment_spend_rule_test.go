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
	runtime_common_mocks "github.com/6022-labs/agentic-agent-node/tests/runtime_common_mocks/services_mocks/interfaces_mocks"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

type WhenCreatingPaymentSpendRuleTestingSuite struct {
	sut *use_cases.CreatePaymentSpendRule

	mockPaymentSpendPolicyProvider       *payment_interfaces_mocks.MockPaymentSpendPolicyProviderInterface
	mockPaymentSpendRuleRepository       *payment_repositories_mocks.MockPaymentSpendRuleRepositoryInterface
	mockPaymentSpendRuleWindowRepository *payment_repositories_mocks.MockPaymentSpendRuleWindowRepositoryInterface
	mockStrategy                         *payment_interfaces_mocks.MockPaymentSpendMethodStrategyInterface
	mockUnitOfWorkFactory                *runtime_common_mocks.MockUnitOfWorkFactoryInterface
	mockUnitOfWork                       *runtime_common_mocks.MockUnitOfWorkInterface
	unitOfWorkCtx                        context.Context
}

func WhenCreatingPaymentSpendRuleBeforeEach(t *testing.T) *WhenCreatingPaymentSpendRuleTestingSuite {
	mockController := gomock.NewController(t)

	mockPaymentSpendPolicyProvider := payment_interfaces_mocks.NewMockPaymentSpendPolicyProviderInterface(
		mockController,
	)
	mockPaymentSpendRuleRepository := payment_repositories_mocks.NewMockPaymentSpendRuleRepositoryInterface(
		mockController,
	)
	mockPaymentSpendRuleWindowRepository :=
		payment_repositories_mocks.NewMockPaymentSpendRuleWindowRepositoryInterface(mockController)
	strategies, mockStrategy := newMockSpendStrategies(mockController)
	mockUnitOfWorkFactory := runtime_common_mocks.NewMockUnitOfWorkFactoryInterface(mockController)
	mockUnitOfWork, unitOfWorkCtx := newMockUnitOfWork(mockController)

	return &WhenCreatingPaymentSpendRuleTestingSuite{
		sut: use_cases.NewCreatePaymentSpendRule(
			mockPaymentSpendPolicyProvider,
			mockPaymentSpendRuleRepository,
			mockPaymentSpendRuleWindowRepository,
			strategies,
			mockUnitOfWorkFactory,
		),

		mockPaymentSpendPolicyProvider:       mockPaymentSpendPolicyProvider,
		mockPaymentSpendRuleRepository:       mockPaymentSpendRuleRepository,
		mockPaymentSpendRuleWindowRepository: mockPaymentSpendRuleWindowRepository,
		mockStrategy:                         mockStrategy,
		mockUnitOfWorkFactory:                mockUnitOfWorkFactory,
		mockUnitOfWork:                       mockUnitOfWork,
		unitOfWorkCtx:                        unitOfWorkCtx,
	}
}

func generateSavePaymentSpendRuleRequest(t *testing.T) *requests.SavePaymentSpendRuleRequest {
	t.Helper()

	config, err := model_builders.NewPaymentSpendRuleX402ConfigBuilder().Build().Marshal()
	assert.NoError(t, err)

	maxAmountPerPayment := "10000"
	return &requests.SavePaymentSpendRuleRequest{
		Name:                "Test token",
		MaxAmountPerPayment: &maxAmountPerPayment,
		Windows:             []requests.SavePaymentSpendRuleWindowRequest{{WindowSeconds: 86400, MaxAmount: "50000"}},
		Config:              config,
	}
}

func TestWhenCreatingPaymentSpendRule(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	t.Run("Given the request is invalid", func(t *testing.T) {
		t.Parallel()

		t.Run("Should return a validation error before touching the policy", func(t *testing.T) {
			t.Parallel()

			suite := WhenCreatingPaymentSpendRuleBeforeEach(t)
			request := generateSavePaymentSpendRuleRequest(t)
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

			suite := WhenCreatingPaymentSpendRuleBeforeEach(t)
			sut := use_cases.NewCreatePaymentSpendRule(
				suite.mockPaymentSpendPolicyProvider,
				suite.mockPaymentSpendRuleRepository,
				suite.mockPaymentSpendRuleWindowRepository,
				newUnsupportedSpendMethodStrategies(gomock.NewController(t)),
				suite.mockUnitOfWorkFactory,
			)

			rule, err := sut.Execute(ctx, generateSavePaymentSpendRuleRequest(t))

			assert.Nil(t, rule)
			var validationErr *apperrors.ValidationError
			assert.ErrorAs(t, err, &validationErr)
			assert.Equal(t, "method", validationErr.Field)
		})
	})

	t.Run("Given the strategy rejects the token", func(t *testing.T) {
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

			rule, err := suite.sut.Execute(ctx, generateSavePaymentSpendRuleRequest(t))

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

			rule, err := suite.sut.Execute(ctx, generateSavePaymentSpendRuleRequest(t))

			assert.Nil(t, rule)
			var internalErr *apperrors.InternalError
			assert.ErrorAs(t, err, &internalErr)
			assert.Equal(t, "ensure_failed", internalErr.Code)
		})
	})

	t.Run("Given looking up the token's rule fails", func(t *testing.T) {
		t.Parallel()

		initSuite := func(suite *WhenCreatingPaymentSpendRuleTestingSuite) {
			policy := model_builders.NewPaymentSpendPolicyBuilder().Build()

			suite.mockStrategy.EXPECT().Validate(gomock.Any(), gomock.Any()).Return(nil).Times(1)
			suite.mockPaymentSpendPolicyProvider.EXPECT().Ensure(gomock.Any()).Return(policy, nil).Times(1)
			suite.mockStrategy.EXPECT().
				RuleIdForAsset(gomock.Any(), policy.Id, gomock.Any()).
				Return(uuid.Nil, assert.AnError).
				Times(1)
		}

		t.Run("Should return an internal error", func(t *testing.T) {
			t.Parallel()

			suite := WhenCreatingPaymentSpendRuleBeforeEach(t)
			initSuite(suite)

			rule, err := suite.sut.Execute(ctx, generateSavePaymentSpendRuleRequest(t))

			assert.Nil(t, rule)
			var internalErr *apperrors.InternalError
			assert.ErrorAs(t, err, &internalErr)
			assert.Equal(t, "rule_lookup_failed", internalErr.Code)
		})
	})

	t.Run("Given another rule already spends the token", func(t *testing.T) {
		t.Parallel()

		initSuite := func(suite *WhenCreatingPaymentSpendRuleTestingSuite) {
			policy := model_builders.NewPaymentSpendPolicyBuilder().Build()

			suite.mockStrategy.EXPECT().Validate(gomock.Any(), gomock.Any()).Return(nil).Times(1)
			suite.mockPaymentSpendPolicyProvider.EXPECT().Ensure(gomock.Any()).Return(policy, nil).Times(1)
			suite.mockStrategy.EXPECT().
				RuleIdForAsset(gomock.Any(), policy.Id, gomock.Any()).
				Return(uuid.New(), nil).
				Times(1)
		}

		t.Run("Should return a conflict error", func(t *testing.T) {
			t.Parallel()

			suite := WhenCreatingPaymentSpendRuleBeforeEach(t)
			initSuite(suite)

			rule, err := suite.sut.Execute(ctx, generateSavePaymentSpendRuleRequest(t))

			assert.Nil(t, rule)
			var conflictErr *apperrors.ConflictError
			assert.ErrorAs(t, err, &conflictErr)
			assert.Equal(t, "asset_already_has_rule", conflictErr.Code)
		})
	})

	t.Run("Given starting the unit of work fails", func(t *testing.T) {
		t.Parallel()

		initSuite := func(suite *WhenCreatingPaymentSpendRuleTestingSuite) {
			policy := model_builders.NewPaymentSpendPolicyBuilder().Build()

			suite.mockStrategy.EXPECT().Validate(gomock.Any(), gomock.Any()).Return(nil).Times(1)
			suite.mockPaymentSpendPolicyProvider.EXPECT().Ensure(gomock.Any()).Return(policy, nil).Times(1)
			suite.mockStrategy.EXPECT().
				RuleIdForAsset(gomock.Any(), policy.Id, gomock.Any()).
				Return(uuid.Nil, nil).
				Times(1)
			suite.mockUnitOfWorkFactory.EXPECT().Start(gomock.Any()).Return(nil, assert.AnError).Times(1)
		}

		t.Run("Should return an internal error", func(t *testing.T) {
			t.Parallel()

			suite := WhenCreatingPaymentSpendRuleBeforeEach(t)
			initSuite(suite)

			rule, err := suite.sut.Execute(ctx, generateSavePaymentSpendRuleRequest(t))

			assert.Nil(t, rule)
			var internalErr *apperrors.InternalError
			assert.ErrorAs(t, err, &internalErr)
			assert.Equal(t, "create_failed", internalErr.Code)
		})
	})

	t.Run("Given the repository fails to persist the rule", func(t *testing.T) {
		t.Parallel()

		initSuite := func(suite *WhenCreatingPaymentSpendRuleTestingSuite) {
			policy := model_builders.NewPaymentSpendPolicyBuilder().Build()

			suite.mockStrategy.EXPECT().Validate(gomock.Any(), gomock.Any()).Return(nil).Times(1)
			suite.mockPaymentSpendPolicyProvider.EXPECT().Ensure(gomock.Any()).Return(policy, nil).Times(1)
			suite.mockStrategy.EXPECT().
				RuleIdForAsset(gomock.Any(), policy.Id, gomock.Any()).
				Return(uuid.Nil, nil).
				Times(1)
			suite.mockUnitOfWorkFactory.EXPECT().Start(gomock.Any()).Return(suite.mockUnitOfWork, nil).Times(1)
			suite.mockPaymentSpendRuleRepository.EXPECT().
				Create(gomock.Any(), gomock.Any()).
				Return(nil, assert.AnError).
				Times(1)
		}

		t.Run("Should return an internal error", func(t *testing.T) {
			t.Parallel()

			suite := WhenCreatingPaymentSpendRuleBeforeEach(t)
			initSuite(suite)

			rule, err := suite.sut.Execute(ctx, generateSavePaymentSpendRuleRequest(t))

			assert.Nil(t, rule)
			var internalErr *apperrors.InternalError
			assert.ErrorAs(t, err, &internalErr)
			assert.Equal(t, "create_failed", internalErr.Code)
		})
	})

	t.Run("Given persisting the token fails", func(t *testing.T) {
		t.Parallel()

		initSuite := func(suite *WhenCreatingPaymentSpendRuleTestingSuite) {
			policy := model_builders.NewPaymentSpendPolicyBuilder().Build()
			rule := model_builders.NewPaymentSpendRuleBuilder().WithPolicyId(policy.Id).Build()

			suite.mockStrategy.EXPECT().Validate(gomock.Any(), gomock.Any()).Return(nil).Times(1)
			suite.mockPaymentSpendPolicyProvider.EXPECT().Ensure(gomock.Any()).Return(policy, nil).Times(1)
			suite.mockStrategy.EXPECT().
				RuleIdForAsset(gomock.Any(), policy.Id, gomock.Any()).
				Return(uuid.Nil, nil).
				Times(1)
			suite.mockUnitOfWorkFactory.EXPECT().Start(gomock.Any()).Return(suite.mockUnitOfWork, nil).Times(1)
			suite.mockPaymentSpendRuleRepository.EXPECT().Create(gomock.Any(), gomock.Any()).Return(rule, nil).Times(1)
			suite.mockStrategy.EXPECT().
				SaveConfig(gomock.Any(), policy.Id, rule.Id, gomock.Any()).
				Return(assert.AnError).
				Times(1)
		}

		t.Run("Should return an internal error", func(t *testing.T) {
			t.Parallel()

			suite := WhenCreatingPaymentSpendRuleBeforeEach(t)
			initSuite(suite)

			rule, err := suite.sut.Execute(ctx, generateSavePaymentSpendRuleRequest(t))

			assert.Nil(t, rule)
			var internalErr *apperrors.InternalError
			assert.ErrorAs(t, err, &internalErr)
			assert.Equal(t, "create_failed", internalErr.Code)
		})
	})

	t.Run("Given persisting a window fails", func(t *testing.T) {
		t.Parallel()

		initSuite := func(suite *WhenCreatingPaymentSpendRuleTestingSuite) {
			policy := model_builders.NewPaymentSpendPolicyBuilder().Build()
			rule := model_builders.NewPaymentSpendRuleBuilder().WithPolicyId(policy.Id).Build()

			suite.mockStrategy.EXPECT().Validate(gomock.Any(), gomock.Any()).Return(nil).Times(1)
			suite.mockPaymentSpendPolicyProvider.EXPECT().Ensure(gomock.Any()).Return(policy, nil).Times(1)
			suite.mockStrategy.EXPECT().
				RuleIdForAsset(gomock.Any(), policy.Id, gomock.Any()).
				Return(uuid.Nil, nil).
				Times(1)
			suite.mockUnitOfWorkFactory.EXPECT().Start(gomock.Any()).Return(suite.mockUnitOfWork, nil).Times(1)
			suite.mockPaymentSpendRuleRepository.EXPECT().Create(gomock.Any(), gomock.Any()).Return(rule, nil).Times(1)
			suite.mockStrategy.EXPECT().SaveConfig(gomock.Any(), policy.Id, rule.Id, gomock.Any()).Return(nil).Times(1)
			suite.mockPaymentSpendRuleWindowRepository.EXPECT().
				Create(gomock.Any(), gomock.Any()).
				Return(nil, assert.AnError).
				Times(1)
		}

		t.Run("Should return an internal error", func(t *testing.T) {
			t.Parallel()

			suite := WhenCreatingPaymentSpendRuleBeforeEach(t)
			initSuite(suite)

			rule, err := suite.sut.Execute(ctx, generateSavePaymentSpendRuleRequest(t))

			assert.Nil(t, rule)
			var internalErr *apperrors.InternalError
			assert.ErrorAs(t, err, &internalErr)
			assert.Equal(t, "create_failed", internalErr.Code)
		})
	})

	t.Run("Given committing the rule fails", func(t *testing.T) {
		t.Parallel()

		initSuite := func(suite *WhenCreatingPaymentSpendRuleTestingSuite) {
			policy := model_builders.NewPaymentSpendPolicyBuilder().Build()
			rule := model_builders.NewPaymentSpendRuleBuilder().WithPolicyId(policy.Id).Build()
			window := model_builders.NewPaymentSpendRuleWindowBuilder().WithPaymentSpendRuleId(rule.Id).Build()

			suite.mockStrategy.EXPECT().Validate(gomock.Any(), gomock.Any()).Return(nil).Times(1)
			suite.mockPaymentSpendPolicyProvider.EXPECT().Ensure(gomock.Any()).Return(policy, nil).Times(1)
			suite.mockStrategy.EXPECT().
				RuleIdForAsset(gomock.Any(), policy.Id, gomock.Any()).
				Return(uuid.Nil, nil).
				Times(1)
			suite.mockUnitOfWorkFactory.EXPECT().Start(gomock.Any()).Return(suite.mockUnitOfWork, nil).Times(1)
			suite.mockPaymentSpendRuleRepository.EXPECT().Create(gomock.Any(), gomock.Any()).Return(rule, nil).Times(1)
			suite.mockStrategy.EXPECT().SaveConfig(gomock.Any(), policy.Id, rule.Id, gomock.Any()).Return(nil).Times(1)
			suite.mockPaymentSpendRuleWindowRepository.EXPECT().
				Create(gomock.Any(), gomock.Any()).
				Return(window, nil).
				Times(1)
			suite.mockUnitOfWork.EXPECT().Commit().Return(assert.AnError).Times(1)
		}

		t.Run("Should return an internal error", func(t *testing.T) {
			t.Parallel()

			suite := WhenCreatingPaymentSpendRuleBeforeEach(t)
			initSuite(suite)

			rule, err := suite.sut.Execute(ctx, generateSavePaymentSpendRuleRequest(t))

			assert.Nil(t, rule)
			var internalErr *apperrors.InternalError
			assert.ErrorAs(t, err, &internalErr)
			assert.Equal(t, "create_failed", internalErr.Code)
		})
	})

	t.Run("Given the rule is valid", func(t *testing.T) {
		t.Parallel()

		policy := model_builders.NewPaymentSpendPolicyBuilder().Build()
		rule := model_builders.NewPaymentSpendRuleBuilder().
			WithPolicyId(policy.Id).
			WithMaxAmountPerPayment("10000").
			Build()
		window := model_builders.NewPaymentSpendRuleWindowBuilder().
			WithPaymentSpendRuleId(rule.Id).
			WithMaxAmount("50000").
			Build()
		maxAmountPerPayment := "10000"

		initSuite := func(suite *WhenCreatingPaymentSpendRuleTestingSuite) {
			suite.mockStrategy.EXPECT().Validate(ctx, gomock.Any()).Return(nil).Times(1)
			suite.mockPaymentSpendPolicyProvider.EXPECT().Ensure(ctx).Return(policy, nil).Times(1)
			suite.mockStrategy.EXPECT().RuleIdForAsset(ctx, policy.Id, gomock.Any()).Return(uuid.Nil, nil).Times(1)
			suite.mockUnitOfWorkFactory.EXPECT().Start(ctx).Return(suite.mockUnitOfWork, nil).Times(1)
		}

		t.Run("Should create the rule on the policy", func(t *testing.T) {
			t.Parallel()

			suite := WhenCreatingPaymentSpendRuleBeforeEach(t)
			initSuite(suite)
			suite.mockPaymentSpendRuleRepository.EXPECT().
				Create(gomock.Any(), &models.PaymentSpendRule{
					PolicyId:            policy.Id,
					Name:                "Test token",
					Method:              enums.PaymentMethodX402,
					MaxAmountPerPayment: &maxAmountPerPayment,
				}).
				Return(rule, nil).
				Times(1)
			suite.mockStrategy.EXPECT().
				SaveConfig(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
				Return(nil).
				Times(1)
			suite.mockPaymentSpendRuleWindowRepository.EXPECT().
				Create(gomock.Any(), gomock.Any()).
				Return(window, nil).
				Times(1)
			suite.mockUnitOfWork.EXPECT().Commit().Return(nil).Times(1)

			_, err := suite.sut.Execute(ctx, generateSavePaymentSpendRuleRequest(t))

			assert.NoError(t, err)
		})

		t.Run("Should persist the token for the rule", func(t *testing.T) {
			t.Parallel()

			suite := WhenCreatingPaymentSpendRuleBeforeEach(t)
			initSuite(suite)
			suite.mockPaymentSpendRuleRepository.EXPECT().Create(gomock.Any(), gomock.Any()).Return(rule, nil).Times(1)
			suite.mockStrategy.EXPECT().SaveConfig(gomock.Any(), policy.Id, rule.Id, gomock.Any()).Return(nil).Times(1)
			suite.mockPaymentSpendRuleWindowRepository.EXPECT().
				Create(gomock.Any(), gomock.Any()).
				Return(window, nil).
				Times(1)
			suite.mockUnitOfWork.EXPECT().Commit().Return(nil).Times(1)

			_, err := suite.sut.Execute(ctx, generateSavePaymentSpendRuleRequest(t))

			assert.NoError(t, err)
		})

		t.Run("Should create each window with nothing spent", func(t *testing.T) {
			t.Parallel()

			suite := WhenCreatingPaymentSpendRuleBeforeEach(t)
			initSuite(suite)
			suite.mockPaymentSpendRuleRepository.EXPECT().Create(gomock.Any(), gomock.Any()).Return(rule, nil).Times(1)
			suite.mockStrategy.EXPECT().
				SaveConfig(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
				Return(nil).
				Times(1)
			suite.mockPaymentSpendRuleWindowRepository.EXPECT().
				Create(gomock.Any(), &models.PaymentSpendRuleWindow{
					PaymentSpendRuleId: rule.Id,
					WindowSeconds:      86400,
					MaxAmount:          "50000",
					Spent:              "0",
				}).
				Return(window, nil).
				Times(1)
			suite.mockUnitOfWork.EXPECT().Commit().Return(nil).Times(1)

			_, err := suite.sut.Execute(ctx, generateSavePaymentSpendRuleRequest(t))

			assert.NoError(t, err)
		})

		t.Run("Should look up the policy and the token outside the unit of work", func(t *testing.T) {
			t.Parallel()

			suite := WhenCreatingPaymentSpendRuleBeforeEach(t)
			initSuite(suite)
			suite.mockPaymentSpendRuleRepository.EXPECT().Create(gomock.Any(), gomock.Any()).Return(rule, nil).Times(1)
			suite.mockStrategy.EXPECT().
				SaveConfig(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
				Return(nil).
				Times(1)
			suite.mockPaymentSpendRuleWindowRepository.EXPECT().
				Create(gomock.Any(), gomock.Any()).
				Return(window, nil).
				Times(1)
			suite.mockUnitOfWork.EXPECT().Commit().Return(nil).Times(1)

			_, err := suite.sut.Execute(ctx, generateSavePaymentSpendRuleRequest(t))

			assert.NoError(t, err)
		})

		t.Run("Should write the rule, its token and its windows inside the unit of work", func(t *testing.T) {
			t.Parallel()

			suite := WhenCreatingPaymentSpendRuleBeforeEach(t)
			initSuite(suite)
			suite.mockPaymentSpendRuleRepository.EXPECT().
				Create(suite.unitOfWorkCtx, gomock.Any()).
				Return(rule, nil).
				Times(1)
			suite.mockStrategy.EXPECT().
				SaveConfig(suite.unitOfWorkCtx, policy.Id, rule.Id, gomock.Any()).
				Return(nil).
				Times(1)
			suite.mockPaymentSpendRuleWindowRepository.EXPECT().
				Create(suite.unitOfWorkCtx, gomock.Any()).
				Return(window, nil).
				Times(1)
			suite.mockUnitOfWork.EXPECT().Commit().Return(nil).Times(1)

			_, err := suite.sut.Execute(ctx, generateSavePaymentSpendRuleRequest(t))

			assert.NoError(t, err)
		})

		t.Run("Should return the rule with its windows", func(t *testing.T) {
			t.Parallel()

			suite := WhenCreatingPaymentSpendRuleBeforeEach(t)
			initSuite(suite)
			suite.mockPaymentSpendRuleRepository.EXPECT().Create(gomock.Any(), gomock.Any()).Return(rule, nil).Times(1)
			suite.mockStrategy.EXPECT().
				SaveConfig(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
				Return(nil).
				Times(1)
			suite.mockPaymentSpendRuleWindowRepository.EXPECT().
				Create(gomock.Any(), gomock.Any()).
				Return(window, nil).
				Times(1)
			suite.mockUnitOfWork.EXPECT().Commit().Return(nil).Times(1)

			response, err := suite.sut.Execute(ctx, generateSavePaymentSpendRuleRequest(t))

			assert.NoError(t, err)
			if assert.Len(t, response.Windows, 1) {
				assert.Equal(t, "50000", response.Windows[0].MaxAmount)
			}
		})
	})
}
