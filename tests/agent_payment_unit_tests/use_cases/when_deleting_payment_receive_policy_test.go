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

type WhenDeletingPaymentReceivePolicyTestingSuite struct {
	sut *use_cases.DeletePaymentReceivePolicy

	mockPaymentReceivePolicyRepository *payment_repositories_mocks.MockPaymentReceivePolicyRepositoryInterface
	mockAgentRepository                *identity_repositories_mocks.MockAgentRepositoryInterface
}

func WhenDeletingPaymentReceivePolicyBeforeEach(t *testing.T) *WhenDeletingPaymentReceivePolicyTestingSuite {
	mockController := gomock.NewController(t)

	mockPaymentReceivePolicyRepository := payment_repositories_mocks.NewMockPaymentReceivePolicyRepositoryInterface(
		mockController,
	)
	mockAgentRepository := identity_repositories_mocks.NewMockAgentRepositoryInterface(mockController)

	return &WhenDeletingPaymentReceivePolicyTestingSuite{
		sut: use_cases.NewDeletePaymentReceivePolicy(mockPaymentReceivePolicyRepository, mockAgentRepository),

		mockPaymentReceivePolicyRepository: mockPaymentReceivePolicyRepository,
		mockAgentRepository:                mockAgentRepository,
	}
}

func TestWhenDeletingPaymentReceivePolicy(t *testing.T) {
	t.Parallel()

	t.Run("Given deleting the policy fails", func(t *testing.T) {
		t.Parallel()

		initSuite := func(suite *WhenDeletingPaymentReceivePolicyTestingSuite) {
			suite.mockAgentRepository.EXPECT().
				Get(gomock.Any()).
				Return(model_builders.NewAgentBuilder().Build(), nil).
				Times(1)
			suite.mockPaymentReceivePolicyRepository.EXPECT().
				DeleteByAgentId(gomock.Any(), gomock.Any()).
				Return(errors.New("db down")).
				Times(1)
		}

		t.Run("Should return an InternalError", func(t *testing.T) {
			t.Parallel()

			suite := WhenDeletingPaymentReceivePolicyBeforeEach(t)
			initSuite(suite)

			err := suite.sut.Execute(context.Background())

			var internalErr *apperrors.InternalError
			assert.ErrorAs(t, err, &internalErr)
		})
	})

	t.Run("Given the policy exists", func(t *testing.T) {
		t.Parallel()

		initSuite := func(suite *WhenDeletingPaymentReceivePolicyTestingSuite) {
			suite.mockAgentRepository.EXPECT().
				Get(gomock.Any()).
				Return(model_builders.NewAgentBuilder().Build(), nil).
				Times(1)
			suite.mockPaymentReceivePolicyRepository.EXPECT().
				DeleteByAgentId(gomock.Any(), gomock.Any()).
				Return(nil).
				Times(1)
		}

		t.Run("Should delete it", func(t *testing.T) {
			t.Parallel()

			suite := WhenDeletingPaymentReceivePolicyBeforeEach(t)
			initSuite(suite)

			err := suite.sut.Execute(context.Background())

			assert.NoError(t, err)
		})
	})
}
