package tools_test

import (
	"context"
	"testing"
	"time"

	os_interfaces "github.com/6022-labs/agentic-agent-node/src/agent_reasoning_os/services/interfaces"
	"github.com/6022-labs/agentic-agent-node/src/agent_reasoning_self_tools/tools"
	"github.com/6022-labs/agentic-agent-node/tests/agent_reasoning_self_tools_mocks/tools_mocks/interfaces_mocks"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

type WhenExecutingListProcessesToolTestingSuite struct {
	sut *tools.ListProcessesTool

	mockListProcessesExecutor *interfaces_mocks.MockListProcessesExecutorInterface
}

func WhenExecutingListProcessesToolBeforeEach(t *testing.T) *WhenExecutingListProcessesToolTestingSuite {
	mockController := gomock.NewController(t)

	mockListProcessesExecutor := interfaces_mocks.NewMockListProcessesExecutorInterface(mockController)

	return &WhenExecutingListProcessesToolTestingSuite{
		sut:                       tools.NewListProcessesTool(mockListProcessesExecutor),
		mockListProcessesExecutor: mockListProcessesExecutor,
	}
}

func TestWhenExecutingListProcessesTool(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	t.Run("Given a running process", func(t *testing.T) {
		t.Parallel()

		initSuite := func(suite *WhenExecutingListProcessesToolTestingSuite) {
			suite.mockListProcessesExecutor.EXPECT().List().Return([]*os_interfaces.BackgroundProcess{
				{Id: uuid.New(), Command: "tail -f log", Running: true, StartedAt: time.Now().UTC()},
			})
		}

		t.Run("Should omit the end date and the exit code", func(t *testing.T) {
			t.Parallel()

			suite := WhenExecutingListProcessesToolBeforeEach(t)
			initSuite(suite)

			out, err := suite.sut.Execute(ctx, nil)

			assert.NoError(t, err)
			// ToToolResult round-trips through json, so entries arrive as []any of map[string]any.
			processes, ok := out["processes"].([]any)
			if assert.True(t, ok) && assert.Len(t, processes, 1) {
				running, _ := processes[0].(map[string]any)
				assert.NotContains(t, running, "ended_at")
				assert.NotContains(t, running, "exit_code")
			}
		})
	})

	t.Run("Given a finished process", func(t *testing.T) {
		t.Parallel()

		initSuite := func(suite *WhenExecutingListProcessesToolTestingSuite) {
			endedAt := time.Now().UTC()
			exitCode := 0
			suite.mockListProcessesExecutor.EXPECT().List().Return([]*os_interfaces.BackgroundProcess{
				{Id: uuid.New(), Command: "true", StartedAt: endedAt, EndedAt: &endedAt, ExitCode: &exitCode},
			})
		}

		t.Run("Should include the end date and the exit code", func(t *testing.T) {
			t.Parallel()

			suite := WhenExecutingListProcessesToolBeforeEach(t)
			initSuite(suite)

			out, err := suite.sut.Execute(ctx, nil)

			assert.NoError(t, err)
			processes, ok := out["processes"].([]any)
			if assert.True(t, ok) && assert.Len(t, processes, 1) {
				finished, _ := processes[0].(map[string]any)
				assert.Contains(t, finished, "ended_at")
				assert.Equal(t, float64(0), finished["exit_code"])
			}
		})
	})

	t.Run("Given no tracked process", func(t *testing.T) {
		t.Parallel()

		t.Run("Should return an object with an empty list", func(t *testing.T) {
			t.Parallel()

			suite := WhenExecutingListProcessesToolBeforeEach(t)
			suite.mockListProcessesExecutor.EXPECT().List().Return(nil)

			out, err := suite.sut.Execute(ctx, nil)

			assert.NoError(t, err)
			assert.Empty(t, out["processes"])
		})
	})
}
