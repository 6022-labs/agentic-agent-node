package use_cases_test

import (
	"context"
	"errors"
	"testing"

	"github.com/6022-labs/agentic-agent-node/src/agent_payment/use_cases"
	apperrors "github.com/6022-labs/agentic-agent-node/src/common/errors"
	identity_repositories_mocks "github.com/6022-labs/agentic-agent-node/tests/agent_identity_mocks/repositories_mocks"
	payment_repositories_mocks "github.com/6022-labs/agentic-agent-node/tests/agent_payment_mocks/repositories_mocks"
	"github.com/6022-labs/agentic-agent-node/tests/common_tests/model_builders"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

type WhenDeletingPaymentSpendPolicyTestingSuite struct {
	sut *use_cases.DeletePaymentSpendPolicy

	mockPaymentSpendPolicyRepository *payment_repositories_mocks.MockPaymentSpendPolicyRepositoryInterface
	mockAgentRepository              *identity_repositories_mocks.MockAgentRepositoryInterface
}

func WhenDeletingPaymentSpendPolicyBeforeEach(t *testing.T) *WhenDeletingPaymentSpendPolicyTestingSuite {
	mockController := gomock.NewController(t)

	mockPaymentSpendPolicyRepository := payment_repositories_mocks.NewMockPaymentSpendPolicyRepositoryInterface(
		mockController,
	)
	mockAgentRepository := identity_repositories_mocks.NewMockAgentRepositoryInterface(mockController)

	return &WhenDeletingPaymentSpendPolicyTestingSuite{
		sut: use_cases.NewDeletePaymentSpendPolicy(mockPaymentSpendPolicyRepository, mockAgentRepository),

		mockPaymentSpendPolicyRepository: mockPaymentSpendPolicyRepository,
		mockAgentRepository:              mockAgentRepository,
	}
}

func TestWhenDeletingPaymentSpendPolicy(t *testing.T) {
	t.Parallel()

	t.Run("Given the agent lookup fails", func(t *testing.T) {
		t.Parallel()

		initSuite := func(suite *WhenDeletingPaymentSpendPolicyTestingSuite) {
			suite.mockAgentRepository.EXPECT().
				Get(gomock.Any()).
				Return(nil, errors.New("db down")).
				Times(1)
		}

		t.Run("Should return an InternalError", func(t *testing.T) {
			t.Parallel()

			suite := WhenDeletingPaymentSpendPolicyBeforeEach(t)
			initSuite(suite)

			err := suite.sut.Execute(context.Background())

			var internalErr *apperrors.InternalError
			if assert.ErrorAs(t, err, &internalErr) {
				assert.Equal(t, "delete_failed", internalErr.Code)
			}
		})
	})

	t.Run("Given the agent identity does not exist yet", func(t *testing.T) {
		t.Parallel()

		initSuite := func(suite *WhenDeletingPaymentSpendPolicyTestingSuite) {
			suite.mockAgentRepository.EXPECT().
				Get(gomock.Any()).
				Return(nil, nil).
				Times(1)
		}

		t.Run("Should succeed without deleting anything", func(t *testing.T) {
			t.Parallel()

			suite := WhenDeletingPaymentSpendPolicyBeforeEach(t)
			initSuite(suite)

			err := suite.sut.Execute(context.Background())

			assert.NoError(t, err)
		})
	})

	t.Run("Given the delete fails", func(t *testing.T) {
		t.Parallel()

		initSuite := func(suite *WhenDeletingPaymentSpendPolicyTestingSuite) {
			suite.mockAgentRepository.EXPECT().
				Get(gomock.Any()).
				Return(model_builders.NewAgentBuilder().Build(), nil).
				Times(1)
			suite.mockPaymentSpendPolicyRepository.EXPECT().
				DeleteByAgentId(gomock.Any(), gomock.Any()).
				Return(errors.New("db down")).
				Times(1)
		}

		t.Run("Should return an InternalError", func(t *testing.T) {
			t.Parallel()

			suite := WhenDeletingPaymentSpendPolicyBeforeEach(t)
			initSuite(suite)

			err := suite.sut.Execute(context.Background())

			var internalErr *apperrors.InternalError
			assert.ErrorAs(t, err, &internalErr)
			assert.Equal(t, "delete_failed", internalErr.Code)
		})
	})

	t.Run("Given the delete succeeds", func(t *testing.T) {
		t.Parallel()

		initSuite := func(suite *WhenDeletingPaymentSpendPolicyTestingSuite) {
			suite.mockAgentRepository.EXPECT().
				Get(gomock.Any()).
				Return(model_builders.NewAgentBuilder().Build(), nil).
				Times(1)
			suite.mockPaymentSpendPolicyRepository.EXPECT().
				DeleteByAgentId(gomock.Any(), gomock.Any()).
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
}
