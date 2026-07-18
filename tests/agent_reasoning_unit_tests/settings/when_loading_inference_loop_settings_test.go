package settings_test

import (
	"testing"

	"github.com/6022-labs/agentic-agent-node/src/agent_reasoning/settings"
	"github.com/knadh/koanf/providers/confmap"
	"github.com/knadh/koanf/v2"
	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
)

func TestWhenLoadingInferenceLoopSettings(t *testing.T) {
	t.Parallel()

	t.Run("Given a configured max iterations", func(t *testing.T) {
		t.Parallel()

		k := koanf.New(".")
		_ = k.Load(confmap.Provider(map[string]any{
			"inference_loop.max_iterations": 25,
		}, "."), nil)

		t.Run("Should load it", func(t *testing.T) {
			t.Parallel()

			result := settings.NewInferenceLoopSettings(zap.NewNop(), k)

			assert.Equal(t, 25, result.MaxIterations)
		})
	})

	t.Run("Given no configuration", func(t *testing.T) {
		t.Parallel()

		k := koanf.New(".")

		t.Run("Should fall back to the default", func(t *testing.T) {
			t.Parallel()

			result := settings.NewInferenceLoopSettings(zap.NewNop(), k)

			assert.Equal(t, 10, result.MaxIterations)
		})
	})

	t.Run("Given a non-positive max iterations", func(t *testing.T) {
		t.Parallel()

		k := koanf.New(".")
		_ = k.Load(confmap.Provider(map[string]any{
			"inference_loop.max_iterations": -1,
		}, "."), nil)

		t.Run("Should fall back to the default", func(t *testing.T) {
			t.Parallel()

			result := settings.NewInferenceLoopSettings(zap.NewNop(), k)

			assert.Equal(t, 10, result.MaxIterations)
		})
	})
}
