package use_cases_test

import (
	"context"
	"testing"

	apperrors "github.com/6022-labs/agentic-agent-node/src/common/errors"
	"github.com/6022-labs/agentic-agent-node/src/gateway/payment/models"
	"github.com/6022-labs/agentic-agent-node/src/gateway/payment/use_cases"
	"github.com/6022-labs/agentic-agent-node/src/gateway/payment/use_cases/requests"
	interfaces_mocks "github.com/6022-labs/agentic-agent-node/tests/gateway_mocks/payment_mocks/services_mocks/interfaces_mocks"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

type WhenUpdatingPaymentSpendPolicyTestingSuite struct {
	sut *use_cases.UpdatePaymentSpendPolicy

	mockPaymentSpendPolicyClient *interfaces_mocks.MockPaymentSpendPolicyClientInterface
}

func WhenUpdatingPaymentSpendPolicyBeforeEach(t *testing.T) *WhenUpdatingPaymentSpendPolicyTestingSuite {
	mockController := gomock.NewController(t)

	mockPaymentSpendPolicyClient := interfaces_mocks.NewMockPaymentSpendPolicyClientInterface(mockController)

	return &WhenUpdatingPaymentSpendPolicyTestingSuite{
		sut: use_cases.NewUpdatePaymentSpendPolicy(mockPaymentSpendPolicyClient),

		mockPaymentSpendPolicyClient: mockPaymentSpendPolicyClient,
	}
}

func TestWhenUpdatingPaymentSpendPolicy(t *testing.T) {
	t.Parallel()

	t.Run("Given the runtime accepts the payees", func(t *testing.T) {
		t.Parallel()

		policy := &models.ExternalPaymentSpendPolicyPayload{
			Id:            "policy-id",
			AgentId:       "agent-id",
			PayToPatterns: []string{"0xabc.*"},
		}

		// The gateway forwards verbatim; the runtime owns validation.
		t.Run("Should forward the payees to the runtime", func(t *testing.T) {
			t.Parallel()

			suite := WhenUpdatingPaymentSpendPolicyBeforeEach(t)
			suite.mockPaymentSpendPolicyClient.EXPECT().
				Update(gomock.Any(), &models.ExternalUpdatePaymentSpendPolicyPayload{
					PayToPatterns: []string{"0xabc.*"},
				}).
				Return(policy, nil).
				Times(1)

			_, err := suite.sut.Execute(context.Background(), &requests.UpdatePaymentSpendPolicyRequest{
				PayToPatterns: []string{"0xabc.*"},
			})

			assert.NoError(t, err)
		})

		t.Run("Should return the runtime's policy", func(t *testing.T) {
			t.Parallel()

			suite := WhenUpdatingPaymentSpendPolicyBeforeEach(t)
			suite.mockPaymentSpendPolicyClient.EXPECT().Update(gomock.Any(), gomock.Any()).Return(policy, nil).Times(1)

			result, err := suite.sut.Execute(context.Background(), &requests.UpdatePaymentSpendPolicyRequest{
				PayToPatterns: []string{"0xabc.*"},
			})

			assert.NoError(t, err)
			assert.Equal(t, "policy-id", result.Id)
			assert.Equal(t, []string{"0xabc.*"}, result.PayToPatterns)
		})
	})

	t.Run("Given the runtime rejects the payees", func(t *testing.T) {
		t.Parallel()

		initSuite := func(suite *WhenUpdatingPaymentSpendPolicyTestingSuite) {
			suite.mockPaymentSpendPolicyClient.EXPECT().
				Update(gomock.Any(), gomock.Any()).
				Return(nil, apperrors.NewValidationError("payToPatterns", "patterns cannot be empty")).
				Times(1)
		}

		t.Run("Should pass the validation error through", func(t *testing.T) {
			t.Parallel()

			suite := WhenUpdatingPaymentSpendPolicyBeforeEach(t)
			initSuite(suite)

			result, err := suite.sut.Execute(context.Background(), &requests.UpdatePaymentSpendPolicyRequest{
				PayToPatterns: []string{" "},
			})

			assert.Nil(t, result)
			var validationErr *apperrors.ValidationError
			assert.ErrorAs(t, err, &validationErr)
			assert.Equal(t, "payToPatterns", validationErr.Field)
		})
	})

	t.Run("Given the runtime is unavailable", func(t *testing.T) {
		t.Parallel()

		initSuite := func(suite *WhenUpdatingPaymentSpendPolicyTestingSuite) {
			suite.mockPaymentSpendPolicyClient.EXPECT().
				Update(gomock.Any(), gomock.Any()).
				Return(nil, apperrors.NewUnavailableError("runtime_unavailable", "connection refused")).
				Times(1)
		}

		t.Run("Should return an unavailable error", func(t *testing.T) {
			t.Parallel()

			suite := WhenUpdatingPaymentSpendPolicyBeforeEach(t)
			initSuite(suite)

			result, err := suite.sut.Execute(context.Background(), &requests.UpdatePaymentSpendPolicyRequest{})

			assert.Nil(t, result)
			var unavailableErr *apperrors.UnavailableError
			assert.ErrorAs(t, err, &unavailableErr)
		})
	})
}
