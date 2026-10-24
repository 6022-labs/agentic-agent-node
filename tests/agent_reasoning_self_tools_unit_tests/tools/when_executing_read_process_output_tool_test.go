package tools_test

import (
	"context"
	"testing"

	os_interfaces "github.com/6022-labs/agentic-agent-node/src/agent_reasoning_os/services/interfaces"
	"github.com/6022-labs/agentic-agent-node/src/agent_reasoning_self_tools/tools"
	"github.com/6022-labs/agentic-agent-node/tests/agent_reasoning_self_tools_mocks/tools_mocks/interfaces_mocks"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

type WhenExecutingReadProcessOutputToolTestingSuite struct {
	sut *tools.ReadProcessOutputTool

	mockReadProcessOutputExecutor *interfaces_mocks.MockReadProcessOutputExecutorInterface
}

func WhenExecutingReadProcessOutputToolBeforeEach(t *testing.T) *WhenExecutingReadProcessOutputToolTestingSuite {
	mockController := gomock.NewController(t)

	mockReadProcessOutputExecutor := interfaces_mocks.NewMockReadProcessOutputExecutorInterface(mockController)

	return &WhenExecutingReadProcessOutputToolTestingSuite{
		sut:                           tools.NewReadProcessOutputTool(mockReadProcessOutputExecutor),
		mockReadProcessOutputExecutor: mockReadProcessOutputExecutor,
	}
}

func TestWhenExecutingReadProcessOutputTool(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	t.Run("Given a malformed session id", func(t *testing.T) {
		t.Parallel()

		t.Run("Should fail without calling the executor", func(t *testing.T) {
			t.Parallel()

			suite := WhenExecutingReadProcessOutputToolBeforeEach(t)

			out, err := suite.sut.Execute(ctx, map[string]any{"session_id": "not-a-uuid"})

			assert.Nil(t, out)
			assert.ErrorContains(t, err, "session_id")
		})
	})

	t.Run("Given a session id", func(t *testing.T) {
		t.Parallel()

		t.Run("Should return the output so far without an exit code", func(t *testing.T) {
			t.Parallel()

			suite := WhenExecutingReadProcessOutputToolBeforeEach(t)
			sessionId := uuid.New()
			suite.mockReadProcessOutputExecutor.EXPECT().
				ReadOutput(sessionId).
				Return(&os_interfaces.CommandResult{SessionId: &sessionId, Stdout: "partial", Running: true}, nil)

			out, err := suite.sut.Execute(ctx, map[string]any{"session_id": sessionId.String()})

			assert.NoError(t, err)
			if assert.NotNil(t, out) {
				assert.Equal(t, "partial", out["stdout"])
				assert.Equal(t, sessionId.String(), out["session_id"])
				assert.NotContains(t, out, "exit_code")
			}
		})
	})
}
