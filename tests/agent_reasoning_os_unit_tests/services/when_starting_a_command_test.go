package services_test

import (
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

type WhenStartingACommandTestingSuite struct {
	sut *services.CommandRunner

	workdir         string
	osSettings      *os_settings.OsSettings
	commandSettings *os_settings.CommandSettings
}

func WhenStartingACommandBeforeEach(t *testing.T) *WhenStartingACommandTestingSuite {
	workdir := t.TempDir()

	osSettings := os_settings.NewOsSettings(zap.NewNop(), koanf.New("."))
	commandSettings := os_settings.NewCommandSettings(
		zap.NewNop(),
		koanf.New("."),
		reasoning_settings.NewInferenceLoopSettings(zap.NewNop(), koanf.New(".")),
	)
	osSettings.Workdir = workdir

	runner, err := services.NewCommandRunner(zap.NewNop(), osSettings, commandSettings)
	assert.NoError(t, err)

	return &WhenStartingACommandTestingSuite{
		sut:             runner,
		workdir:         workdir,
		osSettings:      osSettings,
		commandSettings: commandSettings,
	}
}

func (s *WhenStartingACommandTestingSuite) spec(command string, timeout time.Duration) *os_interfaces.CommandSpec {
	return &os_interfaces.CommandSpec{Command: command, Workdir: s.workdir, Timeout: timeout}
}

func TestWhenStartingACommand(t *testing.T) {
	t.Parallel()

	t.Run("Given a background command", func(t *testing.T) {
		t.Parallel()

		t.Run("Should report its output so far", func(t *testing.T) {
			t.Parallel()

			suite := WhenStartingACommandBeforeEach(t)

			handle, err := suite.sut.Start(suite.spec("echo first; sleep 30", time.Minute))
			assert.NoError(t, err)
			t.Cleanup(func() { _ = handle.Kill() })

			assert.Eventually(t, func() bool {
				return handle.Snapshot().Stdout == "first\n"
			}, 5*time.Second, 20*time.Millisecond)
			assert.True(t, handle.Snapshot().Running)
		})
	})

	t.Run("Given a background command that gets killed", func(t *testing.T) {
		t.Parallel()

		t.Run("Should stop running without timing out", func(t *testing.T) {
			t.Parallel()

			suite := WhenStartingACommandBeforeEach(t)

			handle, err := suite.sut.Start(suite.spec("echo first; sleep 30", time.Minute))
			assert.NoError(t, err)

			assert.NoError(t, handle.Kill())
			// Wider than the runner's 5s pipe-drain delay: a loaded machine hits that path and ties the deadline.
			assert.Eventually(t, func() bool {
				return !handle.Snapshot().Running
			}, 15*time.Second, 20*time.Millisecond)
			assert.False(t, handle.Snapshot().TimedOut)
		})
	})
}
