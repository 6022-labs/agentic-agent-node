package settings_test

import (
	"testing"

	"github.com/6022-labs/agentic-agent-node/src/gateway/core/settings"
	"github.com/knadh/koanf/providers/confmap"
	"github.com/knadh/koanf/v2"
	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
)

func TestWhenLoadingSiweSettings(t *testing.T) {
	t.Parallel()

	consoleSettings := &settings.ConsoleSettings{
		Url:    "https://agent.example.com:8443",
		Domain: "agent.example.com:8443",
	}

	t.Run("Given no configuration", func(t *testing.T) {
		t.Parallel()

		t.Run("Should fall back to the default statement", func(t *testing.T) {
			t.Parallel()

			result := settings.NewSiweSettings(zap.NewNop(), koanf.New("."), consoleSettings)

			assert.Equal(t, "Sign in to your 6022 agent node.", result.Statement)
		})
	})

	t.Run("Given a statement", func(t *testing.T) {
		t.Parallel()

		k := koanf.New(".")
		_ = k.Load(confmap.Provider(map[string]any{"siwe.statement": "Sign in."}, "."), nil)

		t.Run("Should load it and take the uri and domain from the console", func(t *testing.T) {
			t.Parallel()

			result := settings.NewSiweSettings(zap.NewNop(), k, consoleSettings)

			assert.Equal(t, "Sign in.", result.Statement)
			assert.Equal(t, "https://agent.example.com:8443", result.Uri)
			assert.Equal(t, "agent.example.com:8443", result.Domain)
		})
	})
}
