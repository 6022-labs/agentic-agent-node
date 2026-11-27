package settings_test

import (
	"testing"

	"github.com/6022-labs/agentic-agent-node/src/agent_reasoning_http_a2a/settings"
	"github.com/6022-labs/agentic-agent-node/src/common/host_configurations"
	"github.com/knadh/koanf/providers/confmap"
	"github.com/knadh/koanf/v2"
	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
)

func TestWhenLoadingA2ASettings(t *testing.T) {
	t.Run("Given no configuration", func(t *testing.T) {
		t.Run("Should fall back to the defaults", func(t *testing.T) {
			loaded := settings.NewA2ASettings(zap.NewNop(), koanf.New("."))

			assert.Equal(t, 120, loaded.RequestTimeoutSeconds)
			assert.Equal(t, 10, loaded.CardTimeoutSeconds)
			assert.Equal(t, int64(64<<10), loaded.MaxResponseBytes)
			assert.False(t, loaded.AllowPrivateNetworks)
		})
	})

	t.Run("Given configured values", func(t *testing.T) {
		t.Run("Should take them", func(t *testing.T) {
			k := koanf.New(".")
			_ = k.Load(confmap.Provider(map[string]any{
				"a2a": map[string]any{
					"request_timeout_seconds": 45,
					"card_timeout_seconds":    5,
					"max_response_bytes":      2048,
					"allow_private_networks":  true,
				},
			}, "."), nil)

			loaded := settings.NewA2ASettings(zap.NewNop(), k)

			assert.Equal(t, 45, loaded.RequestTimeoutSeconds)
			assert.Equal(t, 5, loaded.CardTimeoutSeconds)
			assert.Equal(t, int64(2048), loaded.MaxResponseBytes)
			assert.True(t, loaded.AllowPrivateNetworks)
		})
	})

	t.Run("Given out-of-range values", func(t *testing.T) {
		t.Run("Should correct them rather than refuse to boot", func(t *testing.T) {
			k := koanf.New(".")
			_ = k.Load(confmap.Provider(map[string]any{
				"a2a": map[string]any{
					"request_timeout_seconds": -1,
					"card_timeout_seconds":    0,
					"max_response_bytes":      0,
				},
			}, "."), nil)

			loaded := settings.NewA2ASettings(zap.NewNop(), k)

			assert.Equal(t, 120, loaded.RequestTimeoutSeconds)
			assert.Equal(t, 10, loaded.CardTimeoutSeconds)
			assert.Equal(t, int64(64<<10), loaded.MaxResponseBytes)
		})
	})

	t.Run("Given an env override", func(t *testing.T) {
		t.Setenv("A2A__REQUEST_TIMEOUT_SECONDS", "45")
		t.Setenv("A2A__ALLOW_PRIVATE_NETWORKS", "true")

		k, err := host_configurations.LoadKoanfConfig()
		assert.NoError(t, err)

		t.Run("Should win over the defaults", func(t *testing.T) {
			loaded := settings.NewA2ASettings(zap.NewNop(), k)

			assert.Equal(t, 45, loaded.RequestTimeoutSeconds)
			assert.True(t, loaded.AllowPrivateNetworks)
		})
	})
}
