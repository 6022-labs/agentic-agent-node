package settings_test

import (
	"testing"

	"github.com/6022-labs/agentic-agent-node/src/agent_reasoning/settings"
	"github.com/knadh/koanf/providers/confmap"
	"github.com/knadh/koanf/v2"
	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
)

func TestWhenLoadingConversationTitlerSettings(t *testing.T) {
	t.Parallel()

	t.Run("Given a configured timeout", func(t *testing.T) {
		t.Parallel()

		k := koanf.New(".")
		_ = k.Load(confmap.Provider(map[string]any{
			"conversation_titler.timeout_seconds": 90,
		}, "."), nil)

		t.Run("Should load it", func(t *testing.T) {
			t.Parallel()

			result := settings.NewConversationTitlerSettings(zap.NewNop(), k)

			assert.Equal(t, 90, result.TimeoutSeconds)
		})
	})

	t.Run("Given no configuration", func(t *testing.T) {
		t.Parallel()

		k := koanf.New(".")

		t.Run("Should fall back to the default", func(t *testing.T) {
			t.Parallel()

			result := settings.NewConversationTitlerSettings(zap.NewNop(), k)

			assert.Equal(t, 30, result.TimeoutSeconds)
		})
	})

	t.Run("Given a non-positive timeout", func(t *testing.T) {
		t.Parallel()

		k := koanf.New(".")
		_ = k.Load(confmap.Provider(map[string]any{
			"conversation_titler.timeout_seconds": 0,
		}, "."), nil)

		t.Run("Should fall back to the default rather than never time out", func(t *testing.T) {
			t.Parallel()

			result := settings.NewConversationTitlerSettings(zap.NewNop(), k)

			assert.Equal(t, 30, result.TimeoutSeconds)
		})
	})
}
