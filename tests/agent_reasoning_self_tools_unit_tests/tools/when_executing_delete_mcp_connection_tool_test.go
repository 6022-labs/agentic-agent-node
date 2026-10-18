package tools_test

import (
	"context"
	"errors"
	"testing"

	reasoning_requests "github.com/6022-labs/agentic-agent-node/src/agent_reasoning/use_cases/requests"
	"github.com/6022-labs/agentic-agent-node/src/agent_reasoning_self_tools/tools"
	"github.com/6022-labs/agentic-agent-node/tests/agent_reasoning_self_tools_mocks/tools_mocks/interfaces_mocks"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

type WhenExecutingDeleteMCPConnectionToolTestingSuite struct {
	sut *tools.DeleteMCPConnectionTool

	mockDeleteMCPConnectionExecutor *interfaces_mocks.MockDeleteMCPConnectionExecutorInterface
}

func WhenExecutingDeleteMCPConnectionToolBeforeEach(t *testing.T) *WhenExecutingDeleteMCPConnectionToolTestingSuite {
	mockController := gomock.NewController(t)

	mockDeleteMCPConnectionExecutor := interfaces_mocks.NewMockDeleteMCPConnectionExecutorInterface(mockController)

	return &WhenExecutingDeleteMCPConnectionToolTestingSuite{
		sut:                             tools.NewDeleteMCPConnectionTool(mockDeleteMCPConnectionExecutor),
		mockDeleteMCPConnectionExecutor: mockDeleteMCPConnectionExecutor,
	}
}

func TestWhenExecutingDeleteMCPConnectionTool(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	t.Run("Given a valid id argument", func(t *testing.T) {
		t.Parallel()

		id := uuid.New()

		t.Run("Should return a success payload", func(t *testing.T) {
			t.Parallel()

			suite := WhenExecutingDeleteMCPConnectionToolBeforeEach(t)
			suite.mockDeleteMCPConnectionExecutor.EXPECT().
				Execute(ctx, &reasoning_requests.DeleteMCPConnectionRequest{Id: id}).
				Return(nil)

			out, err := suite.sut.Execute(ctx, map[string]any{"id": id.String()})

			assert.NoError(t, err)
			assert.Equal(t, map[string]any{"success": true}, out)
		})
	})

	t.Run("Given a malformed id argument", func(t *testing.T) {
		t.Parallel()

		t.Run("Should fail without calling the use case", func(t *testing.T) {
			t.Parallel()

			suite := WhenExecutingDeleteMCPConnectionToolBeforeEach(t)

			out, err := suite.sut.Execute(ctx, map[string]any{"id": "not-a-uuid"})

			assert.Nil(t, out)
			assert.ErrorContains(t, err, "uuid")
		})
	})

	t.Run("Given the use case fails", func(t *testing.T) {
		t.Parallel()

		id := uuid.New()

		t.Run("Should propagate the error", func(t *testing.T) {
			t.Parallel()

			suite := WhenExecutingDeleteMCPConnectionToolBeforeEach(t)
			suite.mockDeleteMCPConnectionExecutor.EXPECT().
				Execute(ctx, &reasoning_requests.DeleteMCPConnectionRequest{Id: id}).
				Return(errors.New("not found"))

			out, err := suite.sut.Execute(ctx, map[string]any{"id": id.String()})

			assert.Nil(t, out)
			assert.ErrorContains(t, err, "not found")
		})
	})
}
