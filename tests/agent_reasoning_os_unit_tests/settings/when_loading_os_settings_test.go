package settings_test

import (
	"testing"

	"github.com/6022-labs/agentic-agent-node/src/agent_reasoning_os/settings"
	"github.com/6022-labs/agentic-agent-node/src/common/host_configurations"
	"github.com/knadh/koanf/providers/confmap"
	"github.com/knadh/koanf/v2"
	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
)

func TestWhenLoadingOsSettings(t *testing.T) {
	t.Run("Given a configured workdir", func(t *testing.T) {
		k := koanf.New(".")
		_ = k.Load(confmap.Provider(map[string]any{
			"os.workdir": "/srv/agent",
		}, "."), nil)

		t.Run("Should load it", func(t *testing.T) {
			result := settings.NewOsSettings(zap.NewNop(), k)

			assert.Equal(t, "/srv/agent", result.Workdir)
		})
	})

	t.Run("Given nothing configured", func(t *testing.T) {
		t.Run("Should fall back to the default", func(t *testing.T) {
			result := settings.NewOsSettings(zap.NewNop(), koanf.New("."))

			assert.Equal(t, "/workspace", result.Workdir)
		})
	})

	t.Run("Given a blank workdir", func(t *testing.T) {
		k := koanf.New(".")
		_ = k.Load(confmap.Provider(map[string]any{
			"os.workdir": "   ",
		}, "."), nil)

		t.Run("Should fall back to the default", func(t *testing.T) {
			result := settings.NewOsSettings(zap.NewNop(), k)

			assert.Equal(t, "/workspace", result.Workdir)
		})
	})

	t.Run("Given a configured command user", func(t *testing.T) {
		k := koanf.New(".")
		_ = k.Load(confmap.Provider(map[string]any{
			"os.command_user": "  agent  ",
		}, "."), nil)

		t.Run("Should load it trimmed", func(t *testing.T) {
			result := settings.NewOsSettings(zap.NewNop(), k)

			assert.Equal(t, "agent", result.CommandUser)
		})
	})

	t.Run("Given no command user", func(t *testing.T) {
		t.Run("Should stay empty so commands keep the runtime's own user", func(t *testing.T) {
			result := settings.NewOsSettings(zap.NewNop(), koanf.New("."))

			assert.Empty(t, result.CommandUser)
		})
	})

	t.Run("Given an env override", func(t *testing.T) {
		t.Setenv("OS__WORKDIR", "/from/env")

		k, err := host_configurations.LoadKoanfConfig()
		assert.NoError(t, err)

		t.Run("Should win over the defaults", func(t *testing.T) {
			result := settings.NewOsSettings(zap.NewNop(), k)

			assert.Equal(t, "/from/env", result.Workdir)
		})
	})
}
