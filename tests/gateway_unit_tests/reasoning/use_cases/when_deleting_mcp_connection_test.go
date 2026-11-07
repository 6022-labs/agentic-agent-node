package use_cases_test

import (
	"context"
	"testing"

	apperrors "github.com/6022-labs/agentic-agent-node/src/common/errors"
	"github.com/6022-labs/agentic-agent-node/src/gateway/reasoning/use_cases"
	"github.com/6022-labs/agentic-agent-node/src/gateway/reasoning/use_cases/requests"
	interfaces_mocks "github.com/6022-labs/agentic-agent-node/tests/gateway_mocks/reasoning_mocks/services_mocks/interfaces_mocks"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

type WhenDeletingMcpConnectionTestingSuite struct {
	sut *use_cases.DeleteMcpConnection

	mockClient *interfaces_mocks.MockMCPConfigClientInterface
}

func WhenDeletingMcpConnectionBeforeEach(t *testing.T) *WhenDeletingMcpConnectionTestingSuite {
	mockController := gomock.NewController(t)

	mockClient := interfaces_mocks.NewMockMCPConfigClientInterface(mockController)

	sut := use_cases.NewDeleteMcpConnection(mockClient)

	return &WhenDeletingMcpConnectionTestingSuite{
		sut: sut,

		mockClient: mockClient,
	}
}

func TestWhenDeletingMcpConnection(t *testing.T) {
	t.Parallel()

	t.Run("Given the runtime answers 404", func(t *testing.T) {
		t.Parallel()

		runtimeErr := apperrors.NewNotFoundError("not_found", "resource not found")
		initSuite := func(suite *WhenDeletingMcpConnectionTestingSuite) {
			suite.mockClient.EXPECT().Delete(gomock.Any(), gomock.Any()).Return(runtimeErr)
		}

		t.Run("Should surface the NotFoundError unchanged", func(t *testing.T) {
			t.Parallel()

			suite := WhenDeletingMcpConnectionBeforeEach(t)
			initSuite(suite)

			err := suite.sut.Execute(context.Background(), &requests.DeleteMcpConnectionRequest{Id: uuid.New()})

			assert.Same(t, runtimeErr, err)
		})
	})

	t.Run("Given the runtime call fails", func(t *testing.T) {
		t.Parallel()

		runtimeErr := apperrors.NewUnavailableError("runtime_unavailable", "the runtime is unavailable", nil)
		initSuite := func(suite *WhenDeletingMcpConnectionTestingSuite) {
			suite.mockClient.EXPECT().Delete(gomock.Any(), gomock.Any()).Return(runtimeErr)
		}

		t.Run("Should surface the UnavailableError unchanged", func(t *testing.T) {
			t.Parallel()

			suite := WhenDeletingMcpConnectionBeforeEach(t)
			initSuite(suite)

			err := suite.sut.Execute(context.Background(), &requests.DeleteMcpConnectionRequest{Id: uuid.New()})

			assert.Same(t, runtimeErr, err)
		})
	})

	t.Run("Given the runtime deletes the connection", func(t *testing.T) {
		t.Parallel()

		id := uuid.New()
		initSuite := func(suite *WhenDeletingMcpConnectionTestingSuite) {
			suite.mockClient.EXPECT().Delete(gomock.Any(), id).Return(nil)
		}

		t.Run("Should forward the id and return NoContent", func(t *testing.T) {
			t.Parallel()

			suite := WhenDeletingMcpConnectionBeforeEach(t)
			initSuite(suite)

			err := suite.sut.Execute(context.Background(), &requests.DeleteMcpConnectionRequest{Id: id})

			assert.NoError(t, err)
		})
	})
}
