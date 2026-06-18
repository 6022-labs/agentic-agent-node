package settings_test

import (
	"testing"

	settings "github.com/6022-labs/agentic-agent-node/src/agent_reasoning_http_inference/settings"
	"github.com/knadh/koanf/providers/confmap"
	"github.com/knadh/koanf/v2"
	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
)

func TestWhenLoadingOpenAISettings(t *testing.T) {
	t.Parallel()

	t.Run("Given a base url configuration", func(t *testing.T) {
		t.Parallel()

		k := koanf.New(".")
		_ = k.Load(confmap.Provider(map[string]any{
			"openai.base_url": "https://openai.example.com",
		}, "."), nil)

		t.Run("Should load the base url as provided", func(t *testing.T) {
			t.Parallel()

			result := settings.NewOpenAISettings(zap.NewNop(), k)

			assert.NotNil(t, result)
			assert.Equal(t, "https://openai.example.com", result.BaseUrl)
		})
	})
}
