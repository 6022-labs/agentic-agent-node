package settings_test

import (
	"testing"

	reasoning_settings "github.com/6022-labs/agentic-agent-node/src/agent_reasoning/settings"
	"github.com/6022-labs/agentic-agent-node/src/agent_reasoning_os/settings"
	"github.com/knadh/koanf/providers/confmap"
	"github.com/knadh/koanf/v2"
	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
)

// loopSettings supplies the budget the command ceiling is warned against.
func loopSettings() *reasoning_settings.InferenceLoopSettings {
	return reasoning_settings.NewInferenceLoopSettings(zap.NewNop(), koanf.New("."))
}

func TestWhenLoadingCommandSettings(t *testing.T) {
	t.Parallel()

	t.Run("Given configured limits", func(t *testing.T) {
		t.Parallel()

		k := koanf.New(".")
		_ = k.Load(confmap.Provider(map[string]any{
			"command.default_timeout_seconds":    30,
			"command.max_timeout_seconds":        90,
			"command.default_yield_seconds":      5,
			"command.max_output_chars":           4096,
			"command.max_background_processes":   2,
			"command.max_background_seconds":     60,
			"command.finished_retention_seconds": 45,
		}, "."), nil)

		t.Run("Should load them", func(t *testing.T) {
			t.Parallel()

			result := settings.NewCommandSettings(zap.NewNop(), k, loopSettings())

			assert.Equal(t, 30, result.DefaultTimeoutSeconds)
			assert.Equal(t, 90, result.MaxTimeoutSeconds)
			assert.Equal(t, 60, result.MaxBackgroundSeconds)
			assert.Equal(t, 45, result.FinishedRetentionSeconds)
			assert.Equal(t, 5, result.DefaultYieldSeconds)
			assert.Equal(t, 4096, result.MaxOutputChars)
			assert.Equal(t, 2, result.MaxBackgroundProcesses)
		})
	})

	t.Run("Given no configuration", func(t *testing.T) {
		t.Parallel()

		t.Run("Should fall back to the defaults", func(t *testing.T) {
			t.Parallel()

			result := settings.NewCommandSettings(zap.NewNop(), koanf.New("."), loopSettings())

			assert.Equal(t, 120, result.DefaultTimeoutSeconds)
			assert.Equal(t, 240, result.MaxTimeoutSeconds)
			assert.Equal(t, 30, result.DefaultYieldSeconds)
			assert.Equal(t, 1800, result.MaxBackgroundSeconds)
			assert.Equal(t, 600, result.FinishedRetentionSeconds)
			assert.Equal(t, 30_000, result.MaxOutputChars)
			assert.Equal(t, 8, result.MaxBackgroundProcesses)
		})
	})

	t.Run("Given non-positive or blank values", func(t *testing.T) {
		t.Parallel()

		k := koanf.New(".")
		_ = k.Load(confmap.Provider(map[string]any{
			"command.max_output_chars":         0,
			"command.max_background_processes": -1,
		}, "."), nil)

		t.Run("Should replace each with its default", func(t *testing.T) {
			t.Parallel()

			result := settings.NewCommandSettings(zap.NewNop(), k, loopSettings())

			assert.Equal(t, 30_000, result.MaxOutputChars)
			assert.Equal(t, 8, result.MaxBackgroundProcesses)
		})
	})

	t.Run("Given a ceiling below the per-call default", func(t *testing.T) {
		t.Parallel()

		k := koanf.New(".")
		_ = k.Load(confmap.Provider(map[string]any{
			"command.default_timeout_seconds": 300,
			"command.max_timeout_seconds":     60,
		}, "."), nil)

		t.Run(
			"Should raise the ceiling so an unqualified command is not capped below its own default",
			func(t *testing.T) {
				t.Parallel()

				result := settings.NewCommandSettings(zap.NewNop(), k, loopSettings())

				assert.Equal(t, 300, result.MaxTimeoutSeconds)
			},
		)
	})

	t.Run("Given a ceiling at or above the loop budget", func(t *testing.T) {
		t.Parallel()

		k := koanf.New(".")
		_ = k.Load(confmap.Provider(map[string]any{
			"command.max_timeout_seconds": 600,
		}, "."), nil)

		t.Run("Should warn but keep the operator value", func(t *testing.T) {
			t.Parallel()

			result := settings.NewCommandSettings(zap.NewNop(), k, loopSettings())

			assert.Equal(t, 600, result.MaxTimeoutSeconds)
		})
	})

	t.Run("Given auto-backgrounding turned off", func(t *testing.T) {
		t.Parallel()

		k := koanf.New(".")
		_ = k.Load(confmap.Provider(map[string]any{
			"command.default_yield_seconds": 0,
		}, "."), nil)

		t.Run("Should keep the zero rather than treat it as unset", func(t *testing.T) {
			t.Parallel()

			result := settings.NewCommandSettings(zap.NewNop(), k, loopSettings())

			assert.Equal(t, 0, result.DefaultYieldSeconds)
		})
	})

	t.Run("Given a negative yield delay", func(t *testing.T) {
		t.Parallel()

		k := koanf.New(".")
		_ = k.Load(confmap.Provider(map[string]any{
			"command.default_yield_seconds": -5,
		}, "."), nil)

		t.Run("Should replace it with the default", func(t *testing.T) {
			t.Parallel()

			result := settings.NewCommandSettings(zap.NewNop(), k, loopSettings())

			assert.Equal(t, 30, result.DefaultYieldSeconds)
		})
	})
}
