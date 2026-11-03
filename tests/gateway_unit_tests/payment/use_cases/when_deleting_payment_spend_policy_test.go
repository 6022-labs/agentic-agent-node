package use_cases_test

import (
	"context"
	"testing"

	apperrors "github.com/6022-labs/agentic-agent-node/src/common/errors"
	"github.com/6022-labs/agentic-agent-node/src/gateway/payment/use_cases"
	interfaces_mocks "github.com/6022-labs/agentic-agent-node/tests/gateway_mocks/payment_mocks/services_mocks/interfaces_mocks"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

type WhenDeletingPaymentSpendPolicyTestingSuite struct {
	sut *use_cases.DeletePaymentSpendPolicy

	mockPaymentSpendPolicyClient *interfaces_mocks.MockPaymentSpendPolicyClientInterface
}

func WhenDeletingPaymentSpendPolicyBeforeEach(t *testing.T) *WhenDeletingPaymentSpendPolicyTestingSuite {
	mockController := gomock.NewController(t)

	mockPaymentSpendPolicyClient := interfaces_mocks.NewMockPaymentSpendPolicyClientInterface(mockController)

	return &WhenDeletingPaymentSpendPolicyTestingSuite{
		sut: use_cases.NewDeletePaymentSpendPolicy(mockPaymentSpendPolicyClient),

		mockPaymentSpendPolicyClient: mockPaymentSpendPolicyClient,
	}
}

func TestWhenDeletingPaymentSpendPolicy(t *testing.T) {
	t.Parallel()

	t.Run("Given the delete succeeds", func(t *testing.T) {
		t.Parallel()

		initSuite := func(suite *WhenDeletingPaymentSpendPolicyTestingSuite) {
			suite.mockPaymentSpendPolicyClient.EXPECT().
				Delete(gomock.Any()).
				Return(nil).
				Times(1)
		}

		t.Run("Should not return an error", func(t *testing.T) {
			t.Parallel()

			suite := WhenDeletingPaymentSpendPolicyBeforeEach(t)
			initSuite(suite)

			err := suite.sut.Execute(context.Background())

			assert.NoError(t, err)
		})
	})

	t.Run("Given the runtime call fails", func(t *testing.T) {
		t.Parallel()

		initSuite := func(suite *WhenDeletingPaymentSpendPolicyTestingSuite) {
			suite.mockPaymentSpendPolicyClient.EXPECT().
				Delete(gomock.Any()).
				Return(apperrors.NewUnavailableError("runtime_unavailable", "the runtime is unavailable")).
				Times(1)
		}

		t.Run("Should return an unavailable error", func(t *testing.T) {
			t.Parallel()

			suite := WhenDeletingPaymentSpendPolicyBeforeEach(t)
			initSuite(suite)

			err := suite.sut.Execute(context.Background())

			var unavailableErr *apperrors.UnavailableError
			assert.ErrorAs(t, err, &unavailableErr)
			assert.Equal(t, "runtime_unavailable", unavailableErr.Code)
		})
	})

	t.Run("Given the runtime answers with a semantic status", func(t *testing.T) {
		t.Parallel()

		// The adapter maps the status; the use case must not flatten it into a 502.
		initSuite := func(suite *WhenDeletingPaymentSpendPolicyTestingSuite) {
			suite.mockPaymentSpendPolicyClient.EXPECT().
				Delete(gomock.Any()).
				Return(apperrors.NewConflictError("agent_not_minted", "")).
				Times(1)
		}

		t.Run("Should pass the conflict through", func(t *testing.T) {
			t.Parallel()

			suite := WhenDeletingPaymentSpendPolicyBeforeEach(t)
			initSuite(suite)

			err := suite.sut.Execute(context.Background())

			var conflictErr *apperrors.ConflictError
			assert.ErrorAs(t, err, &conflictErr)
			assert.Equal(t, "agent_not_minted", conflictErr.Code)
		})
	})
}
