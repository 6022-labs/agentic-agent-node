package services_test

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	reasoning_settings "github.com/6022-labs/agentic-agent-node/src/agent_reasoning/settings"
	"github.com/6022-labs/agentic-agent-node/src/agent_reasoning_os/services"
	os_interfaces "github.com/6022-labs/agentic-agent-node/src/agent_reasoning_os/services/interfaces"
	os_settings "github.com/6022-labs/agentic-agent-node/src/agent_reasoning_os/settings"
	"github.com/knadh/koanf/v2"
	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
)

type WhenRunningACommandTestingSuite struct {
	sut *services.CommandRunner

	workdir         string
	osSettings      *os_settings.OsSettings
	commandSettings *os_settings.CommandSettings
}

func WhenRunningACommandBeforeEach(t *testing.T) *WhenRunningACommandTestingSuite {
	workdir := t.TempDir()

	osSettings := os_settings.NewOsSettings(zap.NewNop(), koanf.New("."))
	commandSettings := os_settings.NewCommandSettings(
		zap.NewNop(),
		koanf.New("."),
		reasoning_settings.NewInferenceLoopSettings(zap.NewNop(), koanf.New(".")),
	)
	osSettings.Workdir = workdir

	return &WhenRunningACommandTestingSuite{
		sut:             newCommandRunner(t, osSettings, commandSettings),
		workdir:         workdir,
		osSettings:      osSettings,
		commandSettings: commandSettings,
	}
}

func newCommandRunner(
	t *testing.T,
	osSettings *os_settings.OsSettings,
	commandSettings *os_settings.CommandSettings,
) *services.CommandRunner {
	runner, err := services.NewCommandRunner(zap.NewNop(), osSettings, commandSettings)
	assert.NoError(t, err)
	return runner
}

func (s *WhenRunningACommandTestingSuite) spec(command string, timeout time.Duration) *os_interfaces.CommandSpec {
	return &os_interfaces.CommandSpec{Command: command, Workdir: s.workdir, Timeout: timeout}
}

func TestWhenRunningACommand(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	t.Run("Given a command that succeeds", func(t *testing.T) {
		t.Parallel()

		t.Run("Should return its stdout and a zero exit code", func(t *testing.T) {
			t.Parallel()

			suite := WhenRunningACommandBeforeEach(t)

			result, err := suite.sut.Run(ctx, suite.spec("echo hello", time.Second*5))

			assert.NoError(t, err)
			if assert.NotNil(t, result) {
				assert.Equal(t, "hello\n", result.Stdout)
				if assert.NotNil(t, result.ExitCode) {
					assert.Equal(t, 0, *result.ExitCode)
				}
				assert.False(t, result.Running)
				assert.False(t, result.TimedOut)
				assert.False(t, result.Truncated)
			}
		})
	})

	t.Run("Given a command that fails", func(t *testing.T) {
		t.Parallel()

		t.Run("Should return its stderr and exit code", func(t *testing.T) {
			t.Parallel()

			suite := WhenRunningACommandBeforeEach(t)

			result, err := suite.sut.Run(ctx, suite.spec("echo boom >&2; exit 3", time.Second*5))

			assert.NoError(t, err)
			if assert.NotNil(t, result) {
				assert.Equal(t, "boom\n", result.Stderr)
				if assert.NotNil(t, result.ExitCode) {
					assert.Equal(t, 3, *result.ExitCode)
				}
			}
		})
	})

	t.Run("Given the runtime's own environment", func(t *testing.T) {
		t.Parallel()

		t.Run("Should not leak it into the command", func(t *testing.T) {
			t.Parallel()

			// os.Setenv rather than t.Setenv: the latter refuses to run alongside parallel tests.
			assert.NoError(t, os.Setenv("VAULT__TOKEN", "super-secret"))
			t.Cleanup(func() { _ = os.Unsetenv("VAULT__TOKEN") })

			suite := WhenRunningACommandBeforeEach(t)

			result, err := suite.sut.Run(ctx, suite.spec("echo \"[$VAULT__TOKEN]\"", time.Second*5))

			assert.NoError(t, err)
			if assert.NotNil(t, result) {
				assert.Equal(t, "[]\n", result.Stdout)
			}
		})
	})

	t.Run("Given output larger than the cap", func(t *testing.T) {
		t.Parallel()

		t.Run("Should keep both ends and flag the truncation", func(t *testing.T) {
			t.Parallel()

			suite := WhenRunningACommandBeforeEach(t)
			suite.commandSettings.MaxOutputChars = 64

			result, err := suite.sut.Run(
				ctx,
				suite.spec("echo start; head -c 500 /dev/zero | tr '\\0' 'x'; echo; echo end", time.Second*5),
			)

			assert.NoError(t, err)
			if assert.NotNil(t, result) {
				assert.True(t, result.Truncated)
				assert.Less(t, len(result.Stdout), 200)
				assert.True(t, strings.HasPrefix(result.Stdout, "start"))
				assert.True(t, strings.HasSuffix(result.Stdout, "end\n"))
			}
		})

		t.Run("Should not flag output that fits within the cap", func(t *testing.T) {
			t.Parallel()

			suite := WhenRunningACommandBeforeEach(t)
			suite.commandSettings.MaxOutputChars = 64

			// Above max/2 so it reaches the tail, but below max so nothing is actually lost.
			result, err := suite.sut.Run(ctx, suite.spec("head -c 40 /dev/zero | tr '\\0' 'x'", time.Second*5))

			assert.NoError(t, err)
			if assert.NotNil(t, result) {
				assert.False(t, result.Truncated)
				assert.NotContains(t, result.Stdout, "truncated")
				assert.Equal(t, 40, len(result.Stdout))
			}
		})

		t.Run("Should keep what it captured when the output only just crosses the cap", func(t *testing.T) {
			t.Parallel()

			suite := WhenRunningACommandBeforeEach(t)
			suite.commandSettings.MaxOutputChars = 64

			result, err := suite.sut.Run(ctx, suite.spec("head -c 65 /dev/zero | tr '\\0' 'x'", time.Second*5))

			assert.NoError(t, err)
			if assert.NotNil(t, result) {
				// Head and tail keep max/2 each, so a byte over the cap costs a byte, not half the buffer.
				assert.True(t, result.Truncated)
				kept := len(result.Stdout) - len("\n… [output truncated] …\n")
				assert.Equal(t, suite.commandSettings.MaxOutputChars, kept)
			}
		})
	})

	t.Run("Given a command that outruns its timeout", func(t *testing.T) {
		t.Parallel()

		t.Run("Should kill the whole process group", func(t *testing.T) {
			t.Parallel()

			suite := WhenRunningACommandBeforeEach(t)
			marker := filepath.Join(suite.workdir, "orphan")

			result, err := suite.sut.Run(
				ctx,
				// The grandchild outlives the direct child unless the kill targets the group.
				suite.spec(fmt.Sprintf("sh -c 'sleep 1 && touch %s' & sleep 10", marker), 200*time.Millisecond),
			)

			assert.NoError(t, err)
			if assert.NotNil(t, result) {
				assert.True(t, result.TimedOut)
				assert.False(t, result.Running)
			}

			time.Sleep(1500 * time.Millisecond)
			_, statError := os.Stat(marker)
			assert.True(t, os.IsNotExist(statError), "the grandchild survived the timeout kill")
		})
	})

	t.Run("Given a command user that does not exist", func(t *testing.T) {
		t.Parallel()

		t.Run("Should refuse to build a runner rather than silently keep the runtime's privileges", func(t *testing.T) {
			t.Parallel()

			osSettings := os_settings.NewOsSettings(zap.NewNop(), koanf.New("."))
			osSettings.Workdir = t.TempDir()
			osSettings.CommandUser = "no-such-account-here"

			runner, err := services.NewCommandRunner(
				zap.NewNop(),
				osSettings,
				os_settings.NewCommandSettings(
					zap.NewNop(),
					koanf.New("."),
					reasoning_settings.NewInferenceLoopSettings(zap.NewNop(), koanf.New(".")),
				),
			)

			assert.Error(t, err)
			assert.Nil(t, runner)
		})
	})

	t.Run("Given a background command", func(t *testing.T) {
		t.Parallel()

		t.Run("Should report progress while running, then its exit code", func(t *testing.T) {
			t.Parallel()

			suite := WhenRunningACommandBeforeEach(t)

			handle, err := suite.sut.Start(suite.spec("echo first; sleep 30", time.Minute))
			assert.NoError(t, err)

			assert.Eventually(t, func() bool {
				return handle.Snapshot().Stdout == "first\n"
			}, 5*time.Second, 20*time.Millisecond)
			assert.True(t, handle.Snapshot().Running)

			assert.NoError(t, handle.Kill())
			// Wider than the runner's 5s pipe-drain delay: a loaded machine hits that path and ties the deadline.
			assert.Eventually(t, func() bool {
				return !handle.Snapshot().Running
			}, 15*time.Second, 20*time.Millisecond)
			assert.False(t, handle.Snapshot().TimedOut)
		})
	})
}
