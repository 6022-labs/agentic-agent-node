package settings_test

import (
	"testing"

	"github.com/6022-labs/agentic-agent-node/src/gateway_host/settings"
	"github.com/knadh/koanf/providers/confmap"
	"github.com/knadh/koanf/v2"
	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
)

func TestWhenLoadingHostSettings(t *testing.T) {
	t.Parallel()

	t.Run("Given a complete configuration", func(t *testing.T) {
		t.Parallel()

		k := koanf.New(".")
		_ = k.Load(confmap.Provider(map[string]any{
			"host.api_port":         5050,
			"host.app_name":         "AgenticAI-Agent-Node-Test",
			"host.body_limit_bytes": 1024 * 1024,
		}, "."), nil)

		t.Run("Should load every field as provided", func(t *testing.T) {
			t.Parallel()

			result := settings.NewHostSettings(zap.NewNop(), k)

			assert.NotNil(t, result)
			assert.Equal(t, 5050, result.ApiPort)
			assert.Equal(t, "AgenticAI-Agent-Node-Test", result.AppName)
			assert.Equal(t, 1024*1024, result.BodyLimitBytes)
		})
	})

	t.Run("Given api_port is missing or zero", func(t *testing.T) {
		t.Parallel()

		k := koanf.New(".")
		_ = k.Load(confmap.Provider(map[string]any{
			"host.app_name": "AgenticAI-Agent-Node-Test",
		}, "."), nil)

		t.Run("Should default api_port to 5000", func(t *testing.T) {
			t.Parallel()

			result := settings.NewHostSettings(zap.NewNop(), k)

			assert.Equal(t, 5000, result.ApiPort)
		})
	})

	t.Run("Given body_limit_bytes is missing or non-positive", func(t *testing.T) {
		t.Parallel()

		k := koanf.New(".")
		_ = k.Load(confmap.Provider(map[string]any{
			"host.app_name":         "AgenticAI-Agent-Node-Test",
			"host.body_limit_bytes": 0,
		}, "."), nil)

		t.Run("Should default body_limit_bytes to 4MB", func(t *testing.T) {
			t.Parallel()

			result := settings.NewHostSettings(zap.NewNop(), k)

			assert.Equal(t, 4*1024*1024, result.BodyLimitBytes)
		})
	})

	t.Run("Given app_name has surrounding whitespace", func(t *testing.T) {
		t.Parallel()

		k := koanf.New(".")
		_ = k.Load(confmap.Provider(map[string]any{
			"host.app_name": "   trimmed-name   ",
		}, "."), nil)

		t.Run("Should trim app_name", func(t *testing.T) {
			t.Parallel()

			result := settings.NewHostSettings(zap.NewNop(), k)

			assert.Equal(t, "trimmed-name", result.AppName)
		})
	})
}
