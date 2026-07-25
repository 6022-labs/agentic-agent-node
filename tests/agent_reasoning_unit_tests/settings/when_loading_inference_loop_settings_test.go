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

	t.Run("Given configured ceilings", func(t *testing.T) {
		t.Parallel()

		k := koanf.New(".")
		_ = k.Load(confmap.Provider(map[string]any{
			"inference_loop.max_iterations":       25,
			"inference_loop.max_duration_seconds": 120,
			"inference_loop.max_total_tokens":     42000,
		}, "."), nil)

		t.Run("Should load them", func(t *testing.T) {
			t.Parallel()

			result := settings.NewInferenceLoopSettings(zap.NewNop(), k)

			assert.Equal(t, 25, result.MaxIterations)
			assert.Equal(t, 120, result.MaxDurationSeconds)
			assert.Equal(t, 42000, result.MaxTotalTokens)
		})
	})

	t.Run("Given no configuration", func(t *testing.T) {
		t.Parallel()

		k := koanf.New(".")

		t.Run("Should fall back to the defaults", func(t *testing.T) {
			t.Parallel()

			result := settings.NewInferenceLoopSettings(zap.NewNop(), k)

			assert.Equal(t, 10, result.MaxIterations)
			assert.Equal(t, 300, result.MaxDurationSeconds)
			assert.Equal(t, 500_000, result.MaxTotalTokens)
		})
	})

	t.Run("Given non-positive ceilings", func(t *testing.T) {
		t.Parallel()

		k := koanf.New(".")
		_ = k.Load(confmap.Provider(map[string]any{
			"inference_loop.max_iterations":       -1,
			"inference_loop.max_duration_seconds": 0,
			"inference_loop.max_total_tokens":     -5,
		}, "."), nil)

		t.Run("Should fall back to the defaults", func(t *testing.T) {
			t.Parallel()

			result := settings.NewInferenceLoopSettings(zap.NewNop(), k)

			assert.Equal(t, 10, result.MaxIterations)
			assert.Equal(t, 300, result.MaxDurationSeconds)
			assert.Equal(t, 500_000, result.MaxTotalTokens)
		})
	})
}
