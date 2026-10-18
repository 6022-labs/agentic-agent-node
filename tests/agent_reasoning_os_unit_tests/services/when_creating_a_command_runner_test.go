package services_test

import (
	"testing"

	reasoning_settings "github.com/6022-labs/agentic-agent-node/src/agent_reasoning/settings"
	"github.com/6022-labs/agentic-agent-node/src/agent_reasoning_os/services"
	os_settings "github.com/6022-labs/agentic-agent-node/src/agent_reasoning_os/settings"
	"github.com/knadh/koanf/v2"
	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
)

func TestWhenCreatingACommandRunner(t *testing.T) {
	t.Parallel()

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

			assert.ErrorContains(t, err, "no-such-account-here")
			assert.Nil(t, runner)
		})
	})
}
