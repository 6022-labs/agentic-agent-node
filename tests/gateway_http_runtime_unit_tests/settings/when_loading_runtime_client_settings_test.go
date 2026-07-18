package settings_test

import (
	"testing"

	"github.com/6022-labs/agentic-agent-node/src/gateway_http_runtime/settings"
	"github.com/knadh/koanf/providers/confmap"
	"github.com/knadh/koanf/v2"
	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
)

func TestWhenLoadingRuntimeClientSettings(t *testing.T) {
	t.Parallel()

	t.Run("Given a configured base url", func(t *testing.T) {
		t.Parallel()

		k := koanf.New(".")
		_ = k.Load(confmap.Provider(map[string]any{
			"runtime_client.base_url": " http://runtime:5400 ",
		}, "."), nil)

		t.Run("Should load and trim it", func(t *testing.T) {
			t.Parallel()

			result := settings.NewRuntimeClientSettings(zap.NewNop(), k)

			assert.Equal(t, "http://runtime:5400", result.BaseUrl)
		})
	})

	t.Run("Given no configuration", func(t *testing.T) {
		t.Parallel()

		k := koanf.New(".")

		t.Run("Should fall back to the loopback default", func(t *testing.T) {
			t.Parallel()

			result := settings.NewRuntimeClientSettings(zap.NewNop(), k)

			assert.Equal(t, "http://127.0.0.1:5400", result.BaseUrl)
		})
	})

	t.Run("Given a blank base url", func(t *testing.T) {
		t.Parallel()

		k := koanf.New(".")
		_ = k.Load(confmap.Provider(map[string]any{
			"runtime_client.base_url": "   ",
		}, "."), nil)

		t.Run("Should fall back to the loopback default", func(t *testing.T) {
			t.Parallel()

			result := settings.NewRuntimeClientSettings(zap.NewNop(), k)

			assert.Equal(t, "http://127.0.0.1:5400", result.BaseUrl)
		})
	})
}
