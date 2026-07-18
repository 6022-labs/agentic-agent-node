package settings_test

import (
	"testing"

	"github.com/6022-labs/agentic-agent-node/src/common/settings"
	"github.com/knadh/koanf/providers/confmap"
	"github.com/knadh/koanf/v2"
	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
)

func TestWhenLoadingTelemetrySettings(t *testing.T) {
	t.Parallel()

	t.Run("Given telemetry enabled with an endpoint and node id", func(t *testing.T) {
		t.Parallel()

		k := koanf.New(".")
		_ = k.Load(confmap.Provider(map[string]any{
			"telemetry.enabled":  true,
			"telemetry.endpoint": " http://collector:4317 ",
			"telemetry.insecure": true,
			"telemetry.node_id":  " node-1 ",
		}, "."), nil)

		t.Run("Should load and trim the values", func(t *testing.T) {
			t.Parallel()

			result := settings.NewTelemetrySettings(zap.NewNop(), k)

			assert.True(t, result.Enabled)
			assert.Equal(t, "http://collector:4317", result.Endpoint)
			assert.True(t, result.Insecure)
			assert.Equal(t, "node-1", result.NodeId)
		})
	})

	t.Run("Given no configuration", func(t *testing.T) {
		t.Parallel()

		k := koanf.New(".")

		t.Run("Should stay disabled", func(t *testing.T) {
			t.Parallel()

			result := settings.NewTelemetrySettings(zap.NewNop(), k)

			assert.False(t, result.Enabled)
			assert.Empty(t, result.Endpoint)
			assert.Empty(t, result.NodeId)
		})
	})
}
