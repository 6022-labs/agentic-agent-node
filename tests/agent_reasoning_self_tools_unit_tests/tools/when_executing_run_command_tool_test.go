package tools_test

import (
	"context"
	"errors"
	"testing"

	os_interfaces "github.com/6022-labs/agentic-agent-node/src/agent_reasoning_os/services/interfaces"
	"github.com/6022-labs/agentic-agent-node/src/agent_reasoning_self_tools/tools"
	"github.com/6022-labs/agentic-agent-node/tests/agent_reasoning_self_tools_mocks/tools_mocks/interfaces_mocks"
	"github.com/google/uuid"
	"github.com/samber/lo"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

type WhenExecutingRunCommandToolTestingSuite struct {
	sut *tools.RunCommandTool

	mockRunCommandExecutor *interfaces_mocks.MockRunCommandExecutorInterface
}

func WhenExecutingRunCommandToolBeforeEach(t *testing.T) *WhenExecutingRunCommandToolTestingSuite {
	mockController := gomock.NewController(t)

	mockRunCommandExecutor := interfaces_mocks.NewMockRunCommandExecutorInterface(mockController)

	return &WhenExecutingRunCommandToolTestingSuite{
		sut:                    tools.NewRunCommandTool(mockRunCommandExecutor),
		mockRunCommandExecutor: mockRunCommandExecutor,
	}
}

func TestWhenExecutingRunCommandTool(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	t.Run("Given only a command argument", func(t *testing.T) {
		t.Parallel()

		t.Run("Should leave the optional arguments unset", func(t *testing.T) {
			t.Parallel()

			suite := WhenExecutingRunCommandToolBeforeEach(t)
			suite.mockRunCommandExecutor.EXPECT().
				Run(ctx, "ls", "", nil, nil, nil).
				Return(&os_interfaces.CommandResult{Stdout: "web\n", ExitCode: lo.ToPtr(0)}, nil)

			out, err := suite.sut.Execute(ctx, map[string]any{"command": "ls"})

			assert.NoError(t, err)
			if assert.NotNil(t, out) {
				assert.Equal(t, "web\n", out["stdout"])
				assert.Equal(t, float64(0), out["exit_code"])
				assert.NotContains(t, out, "session_id")
			}
		})
	})

	t.Run("Given every argument", func(t *testing.T) {
		t.Parallel()

		t.Run("Should forward them all, and report the session id of a detached command", func(t *testing.T) {
			t.Parallel()

			suite := WhenExecutingRunCommandToolBeforeEach(t)
			sessionId := uuid.New()
			suite.mockRunCommandExecutor.EXPECT().
				Run(ctx, "tail -f log", "/tmp", lo.ToPtr(30), lo.ToPtr(5), lo.ToPtr(true)).
				Return(&os_interfaces.CommandResult{SessionId: &sessionId, Running: true}, nil)

			// Numbers and booleans arrive as json scalars, which is what the models emit.
			out, err := suite.sut.Execute(ctx, map[string]any{
				"command":         "tail -f log",
				"workdir":         "/tmp",
				"timeout_seconds": float64(30),
				"yield_seconds":   float64(5),
				"background":      true,
			})

			assert.NoError(t, err)
			if assert.NotNil(t, out) {
				assert.Equal(t, sessionId.String(), out["session_id"])
				assert.Equal(t, true, out["running"])
				assert.NotContains(t, out, "exit_code")
			}
		})
	})

	t.Run("Given background is explicitly false", func(t *testing.T) {
		t.Parallel()

		t.Run("Should forward it as a value rather than an omission", func(t *testing.T) {
			t.Parallel()

			suite := WhenExecutingRunCommandToolBeforeEach(t)
			suite.mockRunCommandExecutor.EXPECT().
				Run(ctx, "ls", "", nil, nil, lo.ToPtr(false)).
				Return(&os_interfaces.CommandResult{ExitCode: lo.ToPtr(0)}, nil)

			out, err := suite.sut.Execute(ctx, map[string]any{"command": "ls", "background": false})

			assert.NoError(t, err)
			assert.NotNil(t, out)
		})
	})

	t.Run("Given the command argument is missing", func(t *testing.T) {
		t.Parallel()

		t.Run("Should fail without calling the executor", func(t *testing.T) {
			t.Parallel()

			suite := WhenExecutingRunCommandToolBeforeEach(t)

			out, err := suite.sut.Execute(ctx, nil)

			assert.Nil(t, out)
			assert.ErrorContains(t, err, "command")
		})
	})

	t.Run("Given the executor fails", func(t *testing.T) {
		t.Parallel()

		t.Run("Should propagate the error", func(t *testing.T) {
			t.Parallel()

			suite := WhenExecutingRunCommandToolBeforeEach(t)
			suite.mockRunCommandExecutor.EXPECT().
				Run(ctx, gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
				Return(nil, errors.New("fork failed"))

			out, err := suite.sut.Execute(ctx, map[string]any{"command": "ls"})

			assert.Nil(t, out)
			assert.ErrorContains(t, err, "fork failed")
		})
	})

	t.Run("Given the tool definition", func(t *testing.T) {
		t.Parallel()

		t.Run("Should require only the command", func(t *testing.T) {
			t.Parallel()

			suite := WhenExecutingRunCommandToolBeforeEach(t)

			suite.mockRunCommandExecutor.EXPECT().
				Defaults().
				Return(&os_interfaces.CommandDefaults{}).
				Times(1)

			definition := suite.sut.Definition()

			assert.Equal(t, "run_command", definition.Name)
			assert.Equal(t, []string{"command"}, definition.InputScheme.Required)
			assert.Len(t, definition.InputScheme.Properties, 5)
		})

		t.Run("Should describe the configured defaults instead of hardcoded ones", func(t *testing.T) {
			t.Parallel()

			suite := WhenExecutingRunCommandToolBeforeEach(t)

			suite.mockRunCommandExecutor.EXPECT().
				Defaults().
				Return(&os_interfaces.CommandDefaults{
					Workdir:              "/srv/agent",
					TimeoutSeconds:       30,
					MaxTimeoutSeconds:    90,
					MaxBackgroundSeconds: 300,
					YieldSeconds:         7,
				}).
				Times(1)

			definition := suite.sut.Definition()

			assert.Contains(t, *definition.InputScheme.Properties["workdir"].Description, "/srv/agent")
			assert.Contains(t, *definition.InputScheme.Properties["timeout_seconds"].Description, "30")
			assert.Contains(t, *definition.InputScheme.Properties["timeout_seconds"].Description, "90")
			assert.Contains(t, *definition.InputScheme.Properties["timeout_seconds"].Description, "300")
			assert.Contains(t, *definition.InputScheme.Properties["yield_seconds"].Description, "7")
		})
	})
}
