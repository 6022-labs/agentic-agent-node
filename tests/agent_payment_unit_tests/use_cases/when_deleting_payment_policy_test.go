package use_cases_test

import (
	"context"
	"errors"
	"testing"

	"github.com/6022-labs/agentic-agent-node/src/agent_payment/use_cases"
	apperrors "github.com/6022-labs/agentic-agent-node/src/common/errors"
	payment_repositories_mocks "github.com/6022-labs/agentic-agent-node/tests/agent_payment_mocks/repositories_mocks"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

type WhenDeletingPaymentPolicyTestingSuite struct {
	sut *use_cases.DeletePaymentPolicy

	mockPaymentPolicyRepository *payment_repositories_mocks.MockPaymentPolicyRepositoryInterface
}

func WhenDeletingPaymentPolicyBeforeEach(t *testing.T) *WhenDeletingPaymentPolicyTestingSuite {
	mockController := gomock.NewController(t)

	mockPaymentPolicyRepository := payment_repositories_mocks.NewMockPaymentPolicyRepositoryInterface(mockController)

	return &WhenDeletingPaymentPolicyTestingSuite{
		sut: use_cases.NewDeletePaymentPolicy(mockPaymentPolicyRepository),

		mockPaymentPolicyRepository: mockPaymentPolicyRepository,
	}
}

func TestWhenDeletingPaymentPolicy(t *testing.T) {
	t.Parallel()

	t.Run("Given the delete succeeds", func(t *testing.T) {
		t.Parallel()

		initSuite := func(suite *WhenDeletingPaymentPolicyTestingSuite) {
			suite.mockPaymentPolicyRepository.EXPECT().
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

	t.Run("Given the delete fails", func(t *testing.T) {
		t.Parallel()

		initSuite := func(suite *WhenDeletingPaymentPolicyTestingSuite) {
			suite.mockPaymentPolicyRepository.EXPECT().
				Delete(gomock.Any()).
				Return(errors.New("db down")).
				Times(1)
		}

		t.Run("Should return an internal error", func(t *testing.T) {
			t.Parallel()

			suite := WhenDeletingPaymentPolicyBeforeEach(t)
			initSuite(suite)

			err := suite.sut.Execute(context.Background())

			var internalErr *apperrors.InternalError
			assert.ErrorAs(t, err, &internalErr)
		})
	})
}
