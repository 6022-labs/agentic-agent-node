package settings_test

import (
	"testing"

	"github.com/6022-labs/agentic-agent-node/src/runtime_host/settings"
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
			"runtime_host.api_port":         6060,
			"runtime_host.body_limit_bytes": 1024 * 1024,
		}, "."), nil)

		t.Run("Should load every field as provided", func(t *testing.T) {
			t.Parallel()

			result := settings.NewHostSettings(zap.NewNop(), k)

			assert.NotNil(t, result)
			assert.Equal(t, 6060, result.ApiPort)
			assert.Equal(t, 1024*1024, result.BodyLimitBytes)
		})
	})

	t.Run("Given api_port is missing or zero", func(t *testing.T) {
		t.Parallel()

		k := koanf.New(".")
		_ = k.Load(confmap.Provider(map[string]any{}, "."), nil)

		t.Run("Should default api_port to 5001", func(t *testing.T) {
			t.Parallel()

			result := settings.NewHostSettings(zap.NewNop(), k)

			assert.Equal(t, 5001, result.ApiPort)
		})
	})

	t.Run("Given body_limit_bytes is missing or non-positive", func(t *testing.T) {
		t.Parallel()

		k := koanf.New(".")
		_ = k.Load(confmap.Provider(map[string]any{
			"runtime_host.body_limit_bytes": 0,
		}, "."), nil)

		t.Run("Should default body_limit_bytes to 4MB", func(t *testing.T) {
			t.Parallel()

			result := settings.NewHostSettings(zap.NewNop(), k)

			assert.Equal(t, 4*1024*1024, result.BodyLimitBytes)
		})
	})

}
