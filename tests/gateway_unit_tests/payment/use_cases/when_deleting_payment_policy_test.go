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

type WhenDeletingPaymentPolicyTestingSuite struct {
	sut *use_cases.DeletePaymentPolicy

	mockPaymentPolicyClient *interfaces_mocks.MockPaymentPolicyClientInterface
}

func WhenDeletingPaymentPolicyBeforeEach(t *testing.T) *WhenDeletingPaymentPolicyTestingSuite {
	mockController := gomock.NewController(t)

	mockPaymentPolicyClient := interfaces_mocks.NewMockPaymentPolicyClientInterface(mockController)

	return &WhenDeletingPaymentPolicyTestingSuite{
		sut: use_cases.NewDeletePaymentPolicy(mockPaymentPolicyClient),

		mockPaymentPolicyClient: mockPaymentPolicyClient,
	}
}

func TestWhenDeletingPaymentPolicy(t *testing.T) {
	t.Parallel()

	t.Run("Given the runtime call fails", func(t *testing.T) {
		t.Parallel()

		initSuite := func(suite *WhenDeletingPaymentPolicyTestingSuite) {
			suite.mockPaymentPolicyClient.EXPECT().
				Delete(gomock.Any()).
				Return(errors.New("connection refused")).
				Times(1)
		}

		t.Run("Should return an UnavailableError", func(t *testing.T) {
			t.Parallel()

			suite := WhenDeletingPaymentPolicyBeforeEach(t)
			initSuite(suite)

			err := suite.sut.Execute(context.Background())

			var unavailableErr *apperrors.UnavailableError
			assert.ErrorAs(t, err, &unavailableErr)
		})
	})

	t.Run("Given the runtime deletes the configuration", func(t *testing.T) {
		t.Parallel()

		initSuite := func(suite *WhenDeletingPaymentPolicyTestingSuite) {
			suite.mockPaymentPolicyClient.EXPECT().
				Delete(gomock.Any()).
				Return(nil).
				Times(1)
		}

		t.Run("Should not return an error", func(t *testing.T) {
			t.Parallel()

			suite := WhenDeletingPaymentPolicyBeforeEach(t)
			initSuite(suite)

			err := suite.sut.Execute(context.Background())

			assert.NoError(t, err)
		})
	})
}
