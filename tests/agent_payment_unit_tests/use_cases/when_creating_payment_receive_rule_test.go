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
	payment_repositories_mocks "github.com/6022-labs/agentic-agent-node/tests/agent_payment_mocks/repositories_mocks"
	payment_interfaces_mocks "github.com/6022-labs/agentic-agent-node/tests/agent_payment_mocks/services_mocks/interfaces_mocks"
	"github.com/6022-labs/agentic-agent-node/tests/common_tests/model_builders"
	runtime_common_mocks "github.com/6022-labs/agentic-agent-node/tests/runtime_common_mocks/services_mocks/interfaces_mocks"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

type WhenCreatingPaymentReceiveRuleTestingSuite struct {
	sut *use_cases.CreatePaymentReceiveRule

	mockPaymentReceivePolicyProvider      *payment_interfaces_mocks.MockPaymentReceivePolicyProviderInterface
	mockPaymentReceiveRuleRepository      *payment_repositories_mocks.MockPaymentReceiveRuleRepositoryInterface
	mockPaymentReceiveRulePriceRepository *payment_repositories_mocks.MockPaymentReceiveRulePriceRepositoryInterface
	mockStrategy                          *payment_interfaces_mocks.MockPaymentReceiveMethodStrategyInterface
	mockUnitOfWorkFactory                 *runtime_common_mocks.MockUnitOfWorkFactoryInterface
	mockUnitOfWork                        *runtime_common_mocks.MockUnitOfWorkInterface
	unitOfWorkCtx                         context.Context
}

func WhenCreatingPaymentReceiveRuleBeforeEach(t *testing.T) *WhenCreatingPaymentReceiveRuleTestingSuite {
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

	return &WhenCreatingPaymentReceiveRuleTestingSuite{
		sut: use_cases.NewCreatePaymentReceiveRule(
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

func generateSavePaymentReceiveRuleRequest() *requests.SavePaymentReceiveRuleRequest {
	return &requests.SavePaymentReceiveRuleRequest{
		SavePaymentRuleRequest: requests.SavePaymentRuleRequest{Name: "Public price"},
		Prices: []requests.SavePaymentReceiveRulePriceRequest{
			{Price: "10000", AddressPatterns: []string{".*"}},
		},
		Config: json.RawMessage(
			`{"network":"eip155:80002","asset":"0x41e94eb019c0762f9bfcf9fb1e58725bfb0e7582",` +
				`"payTo":"0xf39Fd6e51aad88F6F4ce6aB8827279cffFb92266"}`,
		),
	}
}

func TestWhenCreatingPaymentReceiveRule(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	t.Run("Given the request is invalid", func(t *testing.T) {
		t.Parallel()

		t.Run("Should return a validation error before touching the policy", func(t *testing.T) {
			t.Parallel()

			suite := WhenCreatingPaymentReceiveRuleBeforeEach(t)
			request := generateSavePaymentReceiveRuleRequest()
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

			suite := WhenCreatingPaymentReceiveRuleBeforeEach(t)
			sut := use_cases.NewCreatePaymentReceiveRule(
				suite.mockPaymentReceivePolicyProvider,
				suite.mockPaymentReceiveRuleRepository,
				suite.mockPaymentReceiveRulePriceRepository,
				newUnsupportedMethodStrategies(gomock.NewController(t)),
				suite.mockUnitOfWorkFactory,
			)

			rule, err := sut.Execute(ctx, generateSavePaymentReceiveRuleRequest())

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

			rule, err := suite.sut.Execute(ctx, generateSavePaymentReceiveRuleRequest())

			assert.Nil(t, rule)
			var validationErr *apperrors.ValidationError
			assert.ErrorAs(t, err, &validationErr)
			assert.Equal(t, "config.network", validationErr.Field)
		})
	})

	t.Run("Given materializing the policy fails", func(t *testing.T) {
		t.Parallel()

		initSuite := func(suite *WhenCreatingPaymentReceiveRuleTestingSuite) {
			suite.mockStrategy.EXPECT().Validate(gomock.Any(), gomock.Any()).Return(nil).Times(1)
			suite.mockPaymentReceivePolicyProvider.EXPECT().
				Ensure(gomock.Any()).
				Return(nil, nil, apperrors.NewConflictError("agent_not_minted", "")).
				Times(1)
		}

		t.Run("Should propagate the provider's error", func(t *testing.T) {
			t.Parallel()

			suite := WhenCreatingPaymentReceiveRuleBeforeEach(t)
			initSuite(suite)

			rule, err := suite.sut.Execute(ctx, generateSavePaymentReceiveRuleRequest())

			assert.Nil(t, rule)
			var conflictErr *apperrors.ConflictError
			assert.ErrorAs(t, err, &conflictErr)
			assert.Equal(t, "agent_not_minted", conflictErr.Code)
		})
	})

	t.Run("Given looking up the token's rule fails", func(t *testing.T) {
		t.Parallel()

		initSuite := func(suite *WhenCreatingPaymentReceiveRuleTestingSuite) {
			agent := model_builders.NewAgentBuilder().Build()
			policy := model_builders.NewPaymentReceivePolicyBuilder().WithAgentId(agent.Id).Build()

			suite.mockStrategy.EXPECT().Validate(gomock.Any(), gomock.Any()).Return(nil).Times(1)
			suite.mockPaymentReceivePolicyProvider.EXPECT().Ensure(gomock.Any()).Return(policy, agent, nil).Times(1)
			suite.mockStrategy.EXPECT().
				RuleIdForAsset(gomock.Any(), policy.Id, gomock.Any()).
				Return(uuid.Nil, assert.AnError).
				Times(1)
		}

		t.Run("Should return an InternalError", func(t *testing.T) {
			t.Parallel()

			suite := WhenCreatingPaymentReceiveRuleBeforeEach(t)
			initSuite(suite)

			rule, err := suite.sut.Execute(ctx, generateSavePaymentReceiveRuleRequest())

			assert.Nil(t, rule)
			var internalErr *apperrors.InternalError
			assert.ErrorAs(t, err, &internalErr)
			assert.Equal(t, "rule_lookup_failed", internalErr.Code)
		})
	})

	t.Run("Given another rule already prices the token", func(t *testing.T) {
		t.Parallel()

		initSuite := func(suite *WhenCreatingPaymentReceiveRuleTestingSuite) {
			agent := model_builders.NewAgentBuilder().Build()
			policy := model_builders.NewPaymentReceivePolicyBuilder().WithAgentId(agent.Id).Build()

			suite.mockStrategy.EXPECT().Validate(gomock.Any(), gomock.Any()).Return(nil).Times(1)
			suite.mockPaymentReceivePolicyProvider.EXPECT().Ensure(gomock.Any()).Return(policy, agent, nil).Times(1)
			suite.mockStrategy.EXPECT().
				RuleIdForAsset(gomock.Any(), policy.Id, gomock.Any()).
				Return(uuid.New(), nil).
				Times(1)
		}

		t.Run("Should return a conflict error", func(t *testing.T) {
			t.Parallel()

			suite := WhenCreatingPaymentReceiveRuleBeforeEach(t)
			initSuite(suite)

			rule, err := suite.sut.Execute(ctx, generateSavePaymentReceiveRuleRequest())

			assert.Nil(t, rule)
			var conflictErr *apperrors.ConflictError
			assert.ErrorAs(t, err, &conflictErr)
			assert.Equal(t, "asset_already_has_rule", conflictErr.Code)
		})
	})

	t.Run("Given starting the unit of work fails", func(t *testing.T) {
		t.Parallel()

		initSuite := func(suite *WhenCreatingPaymentReceiveRuleTestingSuite) {
			agent := model_builders.NewAgentBuilder().Build()
			policy := model_builders.NewPaymentReceivePolicyBuilder().WithAgentId(agent.Id).Build()

			suite.mockStrategy.EXPECT().Validate(gomock.Any(), gomock.Any()).Return(nil).Times(1)
			suite.mockPaymentReceivePolicyProvider.EXPECT().Ensure(gomock.Any()).Return(policy, agent, nil).Times(1)
			suite.mockStrategy.EXPECT().
				RuleIdForAsset(gomock.Any(), policy.Id, gomock.Any()).
				Return(uuid.Nil, nil).
				Times(1)
			suite.mockUnitOfWorkFactory.EXPECT().Start(gomock.Any()).Return(nil, assert.AnError).Times(1)
		}

		t.Run("Should return an InternalError", func(t *testing.T) {
			t.Parallel()

			suite := WhenCreatingPaymentReceiveRuleBeforeEach(t)
			initSuite(suite)

			rule, err := suite.sut.Execute(ctx, generateSavePaymentReceiveRuleRequest())

			assert.Nil(t, rule)
			var internalErr *apperrors.InternalError
			assert.ErrorAs(t, err, &internalErr)
			assert.Equal(t, "create_failed", internalErr.Code)
		})
	})

	t.Run("Given the repository fails to persist the rule", func(t *testing.T) {
		t.Parallel()

		initSuite := func(suite *WhenCreatingPaymentReceiveRuleTestingSuite) {
			agent := model_builders.NewAgentBuilder().Build()
			policy := model_builders.NewPaymentReceivePolicyBuilder().WithAgentId(agent.Id).Build()

			suite.mockStrategy.EXPECT().Validate(gomock.Any(), gomock.Any()).Return(nil).Times(1)
			suite.mockPaymentReceivePolicyProvider.EXPECT().Ensure(gomock.Any()).Return(policy, agent, nil).Times(1)
			suite.mockStrategy.EXPECT().
				RuleIdForAsset(gomock.Any(), policy.Id, gomock.Any()).
				Return(uuid.Nil, nil).
				Times(1)
			suite.mockUnitOfWorkFactory.EXPECT().Start(gomock.Any()).Return(suite.mockUnitOfWork, nil).Times(1)
			suite.mockPaymentReceiveRuleRepository.EXPECT().
				Create(gomock.Any(), gomock.Any()).
				Return(nil, assert.AnError).
				Times(1)
		}

		t.Run("Should return an InternalError", func(t *testing.T) {
			t.Parallel()

			suite := WhenCreatingPaymentReceiveRuleBeforeEach(t)
			initSuite(suite)

			rule, err := suite.sut.Execute(ctx, generateSavePaymentReceiveRuleRequest())

			assert.Nil(t, rule)
			var internalErr *apperrors.InternalError
			assert.ErrorAs(t, err, &internalErr)
			assert.Equal(t, "create_failed", internalErr.Code)
		})
	})

	t.Run("Given persisting the config fails", func(t *testing.T) {
		t.Parallel()

		initSuite := func(suite *WhenCreatingPaymentReceiveRuleTestingSuite) {
			agent := model_builders.NewAgentBuilder().Build()
			policy := model_builders.NewPaymentReceivePolicyBuilder().WithAgentId(agent.Id).Build()
			rule := model_builders.NewPaymentReceiveRuleBuilder().WithPolicyId(policy.Id).Build()

			suite.mockStrategy.EXPECT().Validate(gomock.Any(), gomock.Any()).Return(nil).Times(1)
			suite.mockPaymentReceivePolicyProvider.EXPECT().Ensure(gomock.Any()).Return(policy, agent, nil).Times(1)
			suite.mockStrategy.EXPECT().
				RuleIdForAsset(gomock.Any(), policy.Id, gomock.Any()).
				Return(uuid.Nil, nil).
				Times(1)
			suite.mockUnitOfWorkFactory.EXPECT().Start(gomock.Any()).Return(suite.mockUnitOfWork, nil).Times(1)
			suite.mockPaymentReceiveRuleRepository.EXPECT().
				Create(gomock.Any(), gomock.Any()).
				Return(rule, nil).
				Times(1)
			suite.mockStrategy.EXPECT().
				SaveConfig(gomock.Any(), policy.Id, rule.Id, gomock.Any()).
				Return(assert.AnError).
				Times(1)
		}

		t.Run("Should return an InternalError", func(t *testing.T) {
			t.Parallel()

			suite := WhenCreatingPaymentReceiveRuleBeforeEach(t)
			initSuite(suite)

			rule, err := suite.sut.Execute(ctx, generateSavePaymentReceiveRuleRequest())

			assert.Nil(t, rule)
			var internalErr *apperrors.InternalError
			assert.ErrorAs(t, err, &internalErr)
			assert.Equal(t, "create_failed", internalErr.Code)
		})
	})

	t.Run("Given persisting a price fails", func(t *testing.T) {
		t.Parallel()

		initSuite := func(suite *WhenCreatingPaymentReceiveRuleTestingSuite) {
			agent := model_builders.NewAgentBuilder().Build()
			policy := model_builders.NewPaymentReceivePolicyBuilder().WithAgentId(agent.Id).Build()
			rule := model_builders.NewPaymentReceiveRuleBuilder().WithPolicyId(policy.Id).Build()

			suite.mockStrategy.EXPECT().Validate(gomock.Any(), gomock.Any()).Return(nil).Times(1)
			suite.mockPaymentReceivePolicyProvider.EXPECT().Ensure(gomock.Any()).Return(policy, agent, nil).Times(1)
			suite.mockStrategy.EXPECT().
				RuleIdForAsset(gomock.Any(), policy.Id, gomock.Any()).
				Return(uuid.Nil, nil).
				Times(1)
			suite.mockUnitOfWorkFactory.EXPECT().Start(gomock.Any()).Return(suite.mockUnitOfWork, nil).Times(1)
			suite.mockPaymentReceiveRuleRepository.EXPECT().
				Create(gomock.Any(), gomock.Any()).
				Return(rule, nil).
				Times(1)
			suite.mockStrategy.EXPECT().SaveConfig(gomock.Any(), policy.Id, rule.Id, gomock.Any()).Return(nil).Times(1)
			suite.mockPaymentReceiveRulePriceRepository.EXPECT().
				Create(gomock.Any(), gomock.Any()).
				Return(nil, assert.AnError).
				Times(1)
		}

		t.Run("Should return an InternalError", func(t *testing.T) {
			t.Parallel()

			suite := WhenCreatingPaymentReceiveRuleBeforeEach(t)
			initSuite(suite)

			rule, err := suite.sut.Execute(ctx, generateSavePaymentReceiveRuleRequest())

			assert.Nil(t, rule)
			var internalErr *apperrors.InternalError
			assert.ErrorAs(t, err, &internalErr)
			assert.Equal(t, "create_failed", internalErr.Code)
		})
	})

	t.Run("Given committing the rule fails", func(t *testing.T) {
		t.Parallel()

		initSuite := func(suite *WhenCreatingPaymentReceiveRuleTestingSuite) {
			agent := model_builders.NewAgentBuilder().Build()
			policy := model_builders.NewPaymentReceivePolicyBuilder().WithAgentId(agent.Id).Build()
			rule := model_builders.NewPaymentReceiveRuleBuilder().WithPolicyId(policy.Id).Build()
			price := model_builders.NewPaymentReceiveRulePriceBuilder().WithPaymentReceiveRuleId(rule.Id).Build()

			suite.mockStrategy.EXPECT().Validate(gomock.Any(), gomock.Any()).Return(nil).Times(1)
			suite.mockPaymentReceivePolicyProvider.EXPECT().Ensure(gomock.Any()).Return(policy, agent, nil).Times(1)
			suite.mockStrategy.EXPECT().
				RuleIdForAsset(gomock.Any(), policy.Id, gomock.Any()).
				Return(uuid.Nil, nil).
				Times(1)
			suite.mockUnitOfWorkFactory.EXPECT().Start(gomock.Any()).Return(suite.mockUnitOfWork, nil).Times(1)
			suite.mockPaymentReceiveRuleRepository.EXPECT().
				Create(gomock.Any(), gomock.Any()).
				Return(rule, nil).
				Times(1)
			suite.mockStrategy.EXPECT().SaveConfig(gomock.Any(), policy.Id, rule.Id, gomock.Any()).Return(nil).Times(1)
			suite.mockPaymentReceiveRulePriceRepository.EXPECT().
				Create(gomock.Any(), gomock.Any()).
				Return(price, nil).
				Times(1)
			suite.mockUnitOfWork.EXPECT().Commit().Return(assert.AnError).Times(1)
		}

		t.Run("Should return an InternalError", func(t *testing.T) {
			t.Parallel()

			suite := WhenCreatingPaymentReceiveRuleBeforeEach(t)
			initSuite(suite)

			rule, err := suite.sut.Execute(ctx, generateSavePaymentReceiveRuleRequest())

			assert.Nil(t, rule)
			var internalErr *apperrors.InternalError
			assert.ErrorAs(t, err, &internalErr)
			assert.Equal(t, "create_failed", internalErr.Code)
		})
	})

	t.Run("Given the rule config is valid", func(t *testing.T) {
		t.Parallel()

		agent := model_builders.NewAgentBuilder().Build()
		policy := model_builders.NewPaymentReceivePolicyBuilder().WithAgentId(agent.Id).Build()
		rule := model_builders.NewPaymentReceiveRuleBuilder().WithPolicyId(policy.Id).WithName("Public price").Build()
		price := model_builders.NewPaymentReceiveRulePriceBuilder().
			WithPaymentReceiveRuleId(rule.Id).
			WithPrice("10000").
			WithAddressPatterns([]string{".*"}).
			Build()

		initSuite := func(suite *WhenCreatingPaymentReceiveRuleTestingSuite) {
			suite.mockStrategy.EXPECT().Validate(ctx, gomock.Any()).Return(nil).Times(1)
			suite.mockPaymentReceivePolicyProvider.EXPECT().Ensure(ctx).Return(policy, agent, nil).Times(1)
			suite.mockStrategy.EXPECT().RuleIdForAsset(ctx, policy.Id, gomock.Any()).Return(uuid.Nil, nil).Times(1)
			suite.mockUnitOfWorkFactory.EXPECT().Start(ctx).Return(suite.mockUnitOfWork, nil).Times(1)
		}

		t.Run("Should create the rule on the policy", func(t *testing.T) {
			t.Parallel()

			suite := WhenCreatingPaymentReceiveRuleBeforeEach(t)
			initSuite(suite)
			suite.mockPaymentReceiveRuleRepository.EXPECT().
				Create(gomock.Any(), &models.PaymentReceiveRule{
					PolicyId: policy.Id,
					Name:     "Public price",
					Method:   enums.PaymentMethodX402,
				}).
				Return(rule, nil).
				Times(1)
			suite.mockStrategy.EXPECT().
				SaveConfig(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
				Return(nil).
				Times(1)
			suite.mockPaymentReceiveRulePriceRepository.EXPECT().
				Create(gomock.Any(), gomock.Any()).
				Return(price, nil).
				Times(1)
			suite.mockUnitOfWork.EXPECT().Commit().Return(nil).Times(1)

			_, err := suite.sut.Execute(ctx, generateSavePaymentReceiveRuleRequest())

			assert.NoError(t, err)
		})

		t.Run("Should persist the config for the rule", func(t *testing.T) {
			t.Parallel()

			suite := WhenCreatingPaymentReceiveRuleBeforeEach(t)
			initSuite(suite)
			suite.mockPaymentReceiveRuleRepository.EXPECT().
				Create(gomock.Any(), gomock.Any()).
				Return(rule, nil).
				Times(1)
			suite.mockStrategy.EXPECT().SaveConfig(gomock.Any(), policy.Id, rule.Id, gomock.Any()).Return(nil).Times(1)
			suite.mockPaymentReceiveRulePriceRepository.EXPECT().
				Create(gomock.Any(), gomock.Any()).
				Return(price, nil).
				Times(1)
			suite.mockUnitOfWork.EXPECT().Commit().Return(nil).Times(1)

			_, err := suite.sut.Execute(ctx, generateSavePaymentReceiveRuleRequest())

			assert.NoError(t, err)
		})

		t.Run("Should create each price with its address patterns", func(t *testing.T) {
			t.Parallel()

			suite := WhenCreatingPaymentReceiveRuleBeforeEach(t)
			initSuite(suite)
			suite.mockPaymentReceiveRuleRepository.EXPECT().
				Create(gomock.Any(), gomock.Any()).
				Return(rule, nil).
				Times(1)
			suite.mockStrategy.EXPECT().
				SaveConfig(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
				Return(nil).
				Times(1)
			suite.mockPaymentReceiveRulePriceRepository.EXPECT().
				Create(gomock.Any(), &models.PaymentReceiveRulePrice{
					PaymentReceiveRuleId: rule.Id,
					Price:                "10000",
					AddressPatterns:      []string{".*"},
				}).
				Return(price, nil).
				Times(1)
			suite.mockUnitOfWork.EXPECT().Commit().Return(nil).Times(1)

			_, err := suite.sut.Execute(ctx, generateSavePaymentReceiveRuleRequest())

			assert.NoError(t, err)
		})

		t.Run("Should save the rule, its config and its prices in one transaction", func(t *testing.T) {
			t.Parallel()

			suite := WhenCreatingPaymentReceiveRuleBeforeEach(t)
			initSuite(suite)
			suite.mockPaymentReceiveRuleRepository.EXPECT().
				Create(suite.unitOfWorkCtx, gomock.Any()).
				Return(rule, nil).
				Times(1)
			suite.mockStrategy.EXPECT().
				SaveConfig(suite.unitOfWorkCtx, policy.Id, rule.Id, gomock.Any()).
				Return(nil).
				Times(1)
			suite.mockPaymentReceiveRulePriceRepository.EXPECT().
				Create(suite.unitOfWorkCtx, gomock.Any()).
				Return(price, nil).
				Times(1)
			suite.mockUnitOfWork.EXPECT().Commit().Return(nil).Times(1)

			_, err := suite.sut.Execute(ctx, generateSavePaymentReceiveRuleRequest())

			assert.NoError(t, err)
		})

		t.Run("Should return the rule with its prices", func(t *testing.T) {
			t.Parallel()

			suite := WhenCreatingPaymentReceiveRuleBeforeEach(t)
			initSuite(suite)
			suite.mockPaymentReceiveRuleRepository.EXPECT().
				Create(gomock.Any(), gomock.Any()).
				Return(rule, nil).
				Times(1)
			suite.mockStrategy.EXPECT().
				SaveConfig(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
				Return(nil).
				Times(1)
			suite.mockPaymentReceiveRulePriceRepository.EXPECT().
				Create(gomock.Any(), gomock.Any()).
				Return(price, nil).
				Times(1)
			suite.mockUnitOfWork.EXPECT().Commit().Return(nil).Times(1)

			response, err := suite.sut.Execute(ctx, generateSavePaymentReceiveRuleRequest())

			assert.NoError(t, err)
			assert.Equal(t, rule.Id, response.Id)
			assert.Equal(t, "Public price", response.Name)
			if assert.Len(t, response.Prices, 1) {
				assert.Equal(t, "10000", response.Prices[0].Price)
				assert.Equal(t, []string{".*"}, response.Prices[0].AddressPatterns)
			}
		})
	})
}
