package use_cases_test

import (
	"context"
	"testing"
	"time"

	"github.com/6022-labs/agentic-agent-node/src/agent_payment/models"
	"github.com/6022-labs/agentic-agent-node/src/agent_payment/use_cases"
	apperrors "github.com/6022-labs/agentic-agent-node/src/common/errors"
	payment_repositories_mocks "github.com/6022-labs/agentic-agent-node/tests/agent_payment_mocks/repositories_mocks"
	payment_interfaces_mocks "github.com/6022-labs/agentic-agent-node/tests/agent_payment_mocks/services_mocks/interfaces_mocks"
	"github.com/6022-labs/agentic-agent-node/tests/common_tests/model_builders"
	runtime_common_mocks "github.com/6022-labs/agentic-agent-node/tests/runtime_common_mocks/services_mocks/interfaces_mocks"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

type WhenListingPaymentSpendRulesTestingSuite struct {
	sut *use_cases.ListPaymentSpendRules

	mockPaymentSpendPolicyProvider       *payment_interfaces_mocks.MockPaymentSpendPolicyProviderInterface
	mockPaymentSpendRuleRepository       *payment_repositories_mocks.MockPaymentSpendRuleRepositoryInterface
	mockPaymentSpendRuleWindowRepository *payment_repositories_mocks.MockPaymentSpendRuleWindowRepositoryInterface
	mockStrategy                         *payment_interfaces_mocks.MockPaymentSpendMethodStrategyInterface
}

func WhenListingPaymentSpendRulesBeforeEach(t *testing.T) *WhenListingPaymentSpendRulesTestingSuite {
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

	return &WhenListingPaymentSpendRulesTestingSuite{
		sut: use_cases.NewListPaymentSpendRules(
			mockPaymentSpendPolicyProvider,
			mockPaymentSpendRuleRepository,
			mockPaymentSpendRuleWindowRepository,
			strategies,
			runtime_common_mocks.NewMockUnitOfWorkFactoryInterface(mockController),
		),

		mockPaymentSpendPolicyProvider:       mockPaymentSpendPolicyProvider,
		mockPaymentSpendRuleRepository:       mockPaymentSpendRuleRepository,
		mockPaymentSpendRuleWindowRepository: mockPaymentSpendRuleWindowRepository,
		mockStrategy:                         mockStrategy,
	}
}

func TestWhenListingPaymentSpendRules(t *testing.T) {
	t.Parallel()

	t.Run("Given materializing the policy fails", func(t *testing.T) {
		t.Parallel()

		initSuite := func(suite *WhenListingPaymentSpendRulesTestingSuite) {
			suite.mockPaymentSpendPolicyProvider.EXPECT().
				Ensure(gomock.Any()).
				Return(nil, apperrors.NewInternalError("ensure_failed", "")).
				Times(1)
		}

		t.Run("Should propagate the provider's error", func(t *testing.T) {
			t.Parallel()

			suite := WhenListingPaymentSpendRulesBeforeEach(t)
			initSuite(suite)

			response, err := suite.sut.Execute(context.Background())

			assert.Nil(t, response)
			var internalErr *apperrors.InternalError
			assert.ErrorAs(t, err, &internalErr)
			assert.Equal(t, "ensure_failed", internalErr.Code)
		})
	})

	t.Run("Given listing the rules fails", func(t *testing.T) {
		t.Parallel()

		initSuite := func(suite *WhenListingPaymentSpendRulesTestingSuite) {
			policy := model_builders.NewPaymentSpendPolicyBuilder().Build()

			suite.mockPaymentSpendPolicyProvider.EXPECT().Ensure(gomock.Any()).Return(policy, nil).Times(1)
			suite.mockPaymentSpendRuleRepository.EXPECT().
				ListByPolicyId(gomock.Any(), policy.Id).
				Return(nil, assert.AnError).
				Times(1)
		}

		t.Run("Should return an internal error", func(t *testing.T) {
			t.Parallel()

			suite := WhenListingPaymentSpendRulesBeforeEach(t)
			initSuite(suite)

			response, err := suite.sut.Execute(context.Background())

			assert.Nil(t, response)
			var internalErr *apperrors.InternalError
			assert.ErrorAs(t, err, &internalErr)
			assert.Equal(t, "list_failed", internalErr.Code)
		})
	})

	t.Run("Given the policy carries no rules", func(t *testing.T) {
		t.Parallel()

		initSuite := func(suite *WhenListingPaymentSpendRulesTestingSuite) {
			policy := model_builders.NewPaymentSpendPolicyBuilder().Build()

			suite.mockPaymentSpendPolicyProvider.EXPECT().Ensure(gomock.Any()).Return(policy, nil).Times(1)
			suite.mockPaymentSpendRuleRepository.EXPECT().
				ListByPolicyId(gomock.Any(), policy.Id).
				Return(nil, nil).
				Times(1)
		}

		t.Run("Should return an empty list rather than nil", func(t *testing.T) {
			t.Parallel()

			suite := WhenListingPaymentSpendRulesBeforeEach(t)
			initSuite(suite)

			response, err := suite.sut.Execute(context.Background())

			assert.NoError(t, err)
			assert.NotNil(t, response.Rules)
		})
	})

	t.Run("Given reading a rule's windows fails", func(t *testing.T) {
		t.Parallel()

		initSuite := func(suite *WhenListingPaymentSpendRulesTestingSuite) {
			policy := model_builders.NewPaymentSpendPolicyBuilder().Build()
			rule := model_builders.NewPaymentSpendRuleBuilder().WithPolicyId(policy.Id).Build()

			suite.mockPaymentSpendPolicyProvider.EXPECT().Ensure(gomock.Any()).Return(policy, nil).Times(1)
			suite.mockPaymentSpendRuleRepository.EXPECT().
				ListByPolicyId(gomock.Any(), policy.Id).
				Return([]models.PaymentSpendRule{*rule}, nil).
				Times(1)
			suite.mockPaymentSpendRuleWindowRepository.EXPECT().
				ListByRuleId(gomock.Any(), rule.Id).
				Return(nil, assert.AnError).
				Times(1)
		}

		t.Run("Should return an internal error", func(t *testing.T) {
			t.Parallel()

			suite := WhenListingPaymentSpendRulesBeforeEach(t)
			initSuite(suite)

			response, err := suite.sut.Execute(context.Background())

			assert.Nil(t, response)
			var internalErr *apperrors.InternalError
			assert.ErrorAs(t, err, &internalErr)
			assert.Equal(t, "list_failed", internalErr.Code)
		})
	})

	t.Run("Given loading a rule's token fails", func(t *testing.T) {
		t.Parallel()

		initSuite := func(suite *WhenListingPaymentSpendRulesTestingSuite) {
			policy := model_builders.NewPaymentSpendPolicyBuilder().Build()
			rule := model_builders.NewPaymentSpendRuleBuilder().WithPolicyId(policy.Id).Build()

			suite.mockPaymentSpendPolicyProvider.EXPECT().Ensure(gomock.Any()).Return(policy, nil).Times(1)
			suite.mockPaymentSpendRuleRepository.EXPECT().
				ListByPolicyId(gomock.Any(), policy.Id).
				Return([]models.PaymentSpendRule{*rule}, nil).
				Times(1)
			suite.mockPaymentSpendRuleWindowRepository.EXPECT().
				ListByRuleId(gomock.Any(), rule.Id).
				Return([]models.PaymentSpendRuleWindow{}, nil).
				Times(1)
			suite.mockStrategy.EXPECT().LoadConfig(gomock.Any(), rule.Id).Return(nil, assert.AnError).Times(1)
		}

		t.Run("Should return an internal error", func(t *testing.T) {
			t.Parallel()

			suite := WhenListingPaymentSpendRulesBeforeEach(t)
			initSuite(suite)

			response, err := suite.sut.Execute(context.Background())

			assert.Nil(t, response)
			var internalErr *apperrors.InternalError
			assert.ErrorAs(t, err, &internalErr)
			assert.Equal(t, "list_failed", internalErr.Code)
		})
	})

	t.Run("Given the policy carries rules", func(t *testing.T) {
		t.Parallel()

		policy := model_builders.NewPaymentSpendPolicyBuilder().Build()
		rule := model_builders.NewPaymentSpendRuleBuilder().WithPolicyId(policy.Id).Build()
		window := model_builders.NewPaymentSpendRuleWindowBuilder().
			WithPaymentSpendRuleId(rule.Id).
			WithSpent("4000").
			WithStartedAt(time.Now().Add(-time.Hour)).
			Build()

		initSuite := func(t *testing.T, suite *WhenListingPaymentSpendRulesTestingSuite) {
			config, err := model_builders.NewPaymentSpendRuleX402ConfigBuilder().Build().Marshal()
			assert.NoError(t, err)

			suite.mockPaymentSpendPolicyProvider.EXPECT().Ensure(gomock.Any()).Return(policy, nil).Times(1)
			suite.mockPaymentSpendRuleRepository.EXPECT().
				ListByPolicyId(gomock.Any(), policy.Id).
				Return([]models.PaymentSpendRule{*rule}, nil).
				Times(1)
			suite.mockPaymentSpendRuleWindowRepository.EXPECT().
				ListByRuleId(gomock.Any(), rule.Id).
				Return([]models.PaymentSpendRuleWindow{*window}, nil).
				Times(1)
			// The token lives in the method's own table, so it is joined back per rule.
			suite.mockStrategy.EXPECT().LoadConfig(gomock.Any(), rule.Id).Return(config, nil).Times(1)
		}

		t.Run("Should return each rule with its token", func(t *testing.T) {
			t.Parallel()

			suite := WhenListingPaymentSpendRulesBeforeEach(t)
			initSuite(t, suite)

			response, err := suite.sut.Execute(context.Background())

			assert.NoError(t, err)
			if assert.Len(t, response.Rules, 1) {
				assert.NotEmpty(t, response.Rules[0].Config)
			}
		})

		t.Run("Should report each window's spend this period", func(t *testing.T) {
			t.Parallel()

			suite := WhenListingPaymentSpendRulesBeforeEach(t)
			initSuite(t, suite)

			response, err := suite.sut.Execute(context.Background())

			assert.NoError(t, err)
			if assert.Len(t, response.Rules, 1) && assert.Len(t, response.Rules[0].Windows, 1) {
				assert.Equal(t, "4000", response.Rules[0].Windows[0].Spent)
			}
		})
	})
}
