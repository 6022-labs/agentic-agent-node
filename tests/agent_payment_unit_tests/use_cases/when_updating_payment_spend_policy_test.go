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
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

type WhenUpdatingPaymentSpendPolicyTestingSuite struct {
	sut *use_cases.UpdatePaymentSpendPolicy

	mockPaymentSpendPolicyProvider   *payment_interfaces_mocks.MockPaymentSpendPolicyProviderInterface
	mockPaymentSpendPolicyRepository *payment_repositories_mocks.MockPaymentSpendPolicyRepositoryInterface
}

func WhenUpdatingPaymentSpendPolicyBeforeEach(t *testing.T) *WhenUpdatingPaymentSpendPolicyTestingSuite {
	mockController := gomock.NewController(t)

	mockPaymentSpendPolicyProvider := payment_interfaces_mocks.NewMockPaymentSpendPolicyProviderInterface(
		mockController,
	)
	mockPaymentSpendPolicyRepository := payment_repositories_mocks.NewMockPaymentSpendPolicyRepositoryInterface(
		mockController,
	)

	return &WhenUpdatingPaymentSpendPolicyTestingSuite{
		sut: use_cases.NewUpdatePaymentSpendPolicy(mockPaymentSpendPolicyProvider, mockPaymentSpendPolicyRepository),

		mockPaymentSpendPolicyProvider:   mockPaymentSpendPolicyProvider,
		mockPaymentSpendPolicyRepository: mockPaymentSpendPolicyRepository,
	}
}

func TestWhenUpdatingPaymentSpendPolicy(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	t.Run("Given the request is invalid", func(t *testing.T) {
		t.Parallel()

		request := &requests.UpdatePaymentSpendPolicyRequest{PayToPatterns: []string{"("}}

		t.Run("Should return a validation error before touching the policy", func(t *testing.T) {
			t.Parallel()

			suite := WhenUpdatingPaymentSpendPolicyBeforeEach(t)

			response, err := suite.sut.Execute(ctx, request)

			assert.Nil(t, response)
			var validationErr *apperrors.ValidationError
			assert.ErrorAs(t, err, &validationErr)
			assert.Equal(t, "payToPatterns", validationErr.Field)
		})
	})

	t.Run("Given materializing the policy fails", func(t *testing.T) {
		t.Parallel()

		initSuite := func(suite *WhenUpdatingPaymentSpendPolicyTestingSuite) {
			suite.mockPaymentSpendPolicyProvider.EXPECT().
				Ensure(gomock.Any()).
				Return(nil, apperrors.NewInternalError("ensure_failed", "")).
				Times(1)
		}

		t.Run("Should propagate the provider's error", func(t *testing.T) {
			t.Parallel()

			suite := WhenUpdatingPaymentSpendPolicyBeforeEach(t)
			initSuite(suite)

			response, err := suite.sut.Execute(
				ctx,
				&requests.UpdatePaymentSpendPolicyRequest{PayToPatterns: []string{".*"}},
			)

			assert.Nil(t, response)
			var internalErr *apperrors.InternalError
			assert.ErrorAs(t, err, &internalErr)
			assert.Equal(t, "ensure_failed", internalErr.Code)
		})
	})

	t.Run("Given persisting the policy fails", func(t *testing.T) {
		t.Parallel()

		initSuite := func(suite *WhenUpdatingPaymentSpendPolicyTestingSuite) {
			policy := model_builders.NewPaymentSpendPolicyBuilder().Build()

			suite.mockPaymentSpendPolicyProvider.EXPECT().Ensure(gomock.Any()).Return(policy, nil).Times(1)
			suite.mockPaymentSpendPolicyRepository.EXPECT().
				Update(gomock.Any(), gomock.Any()).
				Return(nil, assert.AnError).
				Times(1)
		}

		t.Run("Should return an InternalError", func(t *testing.T) {
			t.Parallel()

			suite := WhenUpdatingPaymentSpendPolicyBeforeEach(t)
			initSuite(suite)

			response, err := suite.sut.Execute(
				ctx,
				&requests.UpdatePaymentSpendPolicyRequest{PayToPatterns: []string{".*"}},
			)

			assert.Nil(t, response)
			var internalErr *apperrors.InternalError
			assert.ErrorAs(t, err, &internalErr)
			assert.Equal(t, "update_failed", internalErr.Code)
		})
	})

	t.Run("Given the payees are valid", func(t *testing.T) {
		t.Parallel()

		initSuite := func(suite *WhenUpdatingPaymentSpendPolicyTestingSuite) *models.PaymentSpendPolicy {
			policy := model_builders.NewPaymentSpendPolicyBuilder().WithPayToPatterns([]string{}).Build()

			suite.mockPaymentSpendPolicyProvider.EXPECT().Ensure(gomock.Any()).Return(policy, nil).Times(1)
			return policy
		}

		t.Run("Should replace the payees on the policy", func(t *testing.T) {
			t.Parallel()

			suite := WhenUpdatingPaymentSpendPolicyBeforeEach(t)
			policy := initSuite(suite)
			suite.mockPaymentSpendPolicyRepository.EXPECT().
				Update(gomock.Any(), &models.PaymentSpendPolicy{
					Id:            policy.Id,
					AgentId:       policy.AgentId,
					PayToPatterns: []string{"0xabc.*"},
				}).
				Return(policy, nil).
				Times(1)

			_, err := suite.sut.Execute(
				ctx,
				&requests.UpdatePaymentSpendPolicyRequest{PayToPatterns: []string{"0xabc.*"}},
			)

			assert.NoError(t, err)
		})

		t.Run("Should return the updated payees", func(t *testing.T) {
			t.Parallel()

			suite := WhenUpdatingPaymentSpendPolicyBeforeEach(t)
			policy := initSuite(suite)
			suite.mockPaymentSpendPolicyRepository.EXPECT().
				Update(gomock.Any(), gomock.Any()).
				Return(policy, nil).
				Times(1)

			response, err := suite.sut.Execute(
				ctx,
				&requests.UpdatePaymentSpendPolicyRequest{PayToPatterns: []string{"0xabc.*"}},
			)

			assert.NoError(t, err)
			assert.Equal(t, []string{"0xabc.*"}, response.PayToPatterns)
		})
	})
}
