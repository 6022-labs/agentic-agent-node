package tools_test

import (
	"context"
	"errors"
	"testing"

	"github.com/6022-labs/agentic-agent-node/src/agent_reasoning_self_tools/tools"
	"github.com/6022-labs/agentic-agent-node/tests/agent_reasoning_self_tools_mocks/tools_mocks/interfaces_mocks"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

type WhenExecutingKillProcessToolTestingSuite struct {
	sut *tools.KillProcessTool

	mockKillProcessExecutor *interfaces_mocks.MockKillProcessExecutorInterface
}

func WhenExecutingKillProcessToolBeforeEach(t *testing.T) *WhenExecutingKillProcessToolTestingSuite {
	mockController := gomock.NewController(t)

	mockKillProcessExecutor := interfaces_mocks.NewMockKillProcessExecutorInterface(mockController)

	return &WhenExecutingKillProcessToolTestingSuite{
		sut:                     tools.NewKillProcessTool(mockKillProcessExecutor),
		mockKillProcessExecutor: mockKillProcessExecutor,
	}
}

func TestWhenExecutingKillProcessTool(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	t.Run("Given a session id", func(t *testing.T) {
		t.Parallel()

		t.Run("Should report success", func(t *testing.T) {
			t.Parallel()

			suite := WhenExecutingKillProcessToolBeforeEach(t)
			sessionId := uuid.New()
			suite.mockKillProcessExecutor.EXPECT().Kill(ctx, sessionId).Return(nil)

			out, err := suite.sut.Execute(ctx, map[string]any{"session_id": sessionId.String()})

			assert.NoError(t, err)
			if assert.NotNil(t, out) {
				assert.Equal(t, true, out["success"])
			}
		})
	})

	t.Run("Given the executor fails", func(t *testing.T) {
		t.Parallel()

		t.Run("Should propagate the error", func(t *testing.T) {
			t.Parallel()

			suite := WhenExecutingKillProcessToolBeforeEach(t)
			suite.mockKillProcessExecutor.EXPECT().
				Kill(ctx, gomock.Any()).
				Return(errors.New("no such process"))

			out, err := suite.sut.Execute(ctx, map[string]any{"session_id": uuid.New().String()})

			assert.Nil(t, out)
			assert.ErrorContains(t, err, "no such process")
		})
	})
}
