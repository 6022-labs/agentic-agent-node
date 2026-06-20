package settings_test

import (
	"testing"

	"github.com/6022-labs/agentic-agent-node/src/common/settings"
	"github.com/knadh/koanf/providers/confmap"
	"github.com/knadh/koanf/v2"
	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
)

func TestWhenLoadingHostSettings(t *testing.T) {
	t.Parallel()

	t.Run("Given a complete configuration under the given key", func(t *testing.T) {
		t.Parallel()

		k := koanf.New(".")
		_ = k.Load(confmap.Provider(map[string]any{
			"runtime_host.api_port":         6060,
			"runtime_host.body_limit_bytes": 1024 * 1024,
			"runtime_host.listen_address":   "127.0.0.1",
		}, "."), nil)

		t.Run("Should load every field as provided", func(t *testing.T) {
			t.Parallel()

			result := settings.NewHostSettings(zap.NewNop(), k, "runtime_host", 5400)

			assert.NotNil(t, result)
			assert.Equal(t, 6060, result.ApiPort)
			assert.Equal(t, 1024*1024, result.BodyLimitBytes)
			assert.Equal(t, "127.0.0.1", result.ListenAddress)
		})
	})

	t.Run("Given listen_address is missing", func(t *testing.T) {
		t.Parallel()

		k := koanf.New(".")
		_ = k.Load(confmap.Provider(map[string]any{}, "."), nil)

		t.Run("Should default to empty (all interfaces)", func(t *testing.T) {
			t.Parallel()

			result := settings.NewHostSettings(zap.NewNop(), k, "runtime_host", 5400)

			assert.Equal(t, "", result.ListenAddress)
		})
	})

	t.Run("Given api_port is missing or zero", func(t *testing.T) {
		t.Parallel()

		k := koanf.New(".")
		_ = k.Load(confmap.Provider(map[string]any{}, "."), nil)

		t.Run("Should fall back to the provided default api_port", func(t *testing.T) {
			t.Parallel()

			runtime := settings.NewHostSettings(zap.NewNop(), k, "runtime_host", 5400)
			gateway := settings.NewHostSettings(zap.NewNop(), k, "gateway_host", 5000)

			assert.Equal(t, 5400, runtime.ApiPort)
			assert.Equal(t, 5000, gateway.ApiPort)
		})
	})

	t.Run("Given body_limit_bytes is missing or non-positive", func(t *testing.T) {
		t.Parallel()

		k := koanf.New(".")
		_ = k.Load(confmap.Provider(map[string]any{
			"gateway_host.body_limit_bytes": 0,
		}, "."), nil)

		t.Run("Should default body_limit_bytes to 4MB", func(t *testing.T) {
			t.Parallel()

			result := settings.NewHostSettings(zap.NewNop(), k, "gateway_host", 5000)

			assert.Equal(t, 4*1024*1024, result.BodyLimitBytes)
		})
	})
}
