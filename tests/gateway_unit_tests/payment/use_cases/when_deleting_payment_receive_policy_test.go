package use_cases_test

import (
	"context"
	"errors"
	"testing"

	apperrors "github.com/6022-labs/agentic-agent-node/src/common/errors"
	"github.com/6022-labs/agentic-agent-node/src/gateway/payment/use_cases"
	interfaces_mocks "github.com/6022-labs/agentic-agent-node/tests/gateway_mocks/payment_mocks/services_mocks/interfaces_mocks"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

type WhenDeletingPaymentReceivePolicyTestingSuite struct {
	sut *use_cases.DeletePaymentReceivePolicy

	mockPaymentReceivePolicyClient *interfaces_mocks.MockPaymentReceivePolicyClientInterface
}

func WhenDeletingPaymentReceivePolicyBeforeEach(t *testing.T) *WhenDeletingPaymentReceivePolicyTestingSuite {
	mockController := gomock.NewController(t)

	mockPaymentReceivePolicyClient := interfaces_mocks.NewMockPaymentReceivePolicyClientInterface(mockController)

	return &WhenDeletingPaymentReceivePolicyTestingSuite{
		sut: use_cases.NewDeletePaymentReceivePolicy(mockPaymentReceivePolicyClient),

		mockPaymentReceivePolicyClient: mockPaymentReceivePolicyClient,
	}
}

func TestWhenDeletingPaymentReceivePolicy(t *testing.T) {
	t.Parallel()

	t.Run("Given the runtime call fails", func(t *testing.T) {
		t.Parallel()

		initSuite := func(suite *WhenDeletingPaymentReceivePolicyTestingSuite) {
			suite.mockPaymentReceivePolicyClient.EXPECT().
				Delete(gomock.Any()).
				Return(errors.New("connection refused")).
				Times(1)
		}

		t.Run("Should return an UnavailableError", func(t *testing.T) {
			t.Parallel()

			suite := WhenDeletingPaymentReceivePolicyBeforeEach(t)
			initSuite(suite)

			err := suite.sut.Execute(context.Background())

			var unavailableErr *apperrors.UnavailableError
			assert.ErrorAs(t, err, &unavailableErr)
		})
	})

	t.Run("Given the runtime deletes the configuration", func(t *testing.T) {
		t.Parallel()

		initSuite := func(suite *WhenDeletingPaymentReceivePolicyTestingSuite) {
			suite.mockPaymentReceivePolicyClient.EXPECT().
				Delete(gomock.Any()).
				Return(nil).
				Times(1)
		}

		t.Run("Should not return an error", func(t *testing.T) {
			t.Parallel()

			suite := WhenDeletingPaymentReceivePolicyBeforeEach(t)
			initSuite(suite)

			err := suite.sut.Execute(context.Background())

			assert.NoError(t, err)
		})
	})
}
