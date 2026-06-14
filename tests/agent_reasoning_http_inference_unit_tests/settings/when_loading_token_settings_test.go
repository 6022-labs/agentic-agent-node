package settings_test

import (
	"testing"

	settings "github.com/6022-labs/agentic-agent-node/src/agent_reasoning_http_inference/settings"
	"github.com/knadh/koanf/providers/confmap"
	"github.com/knadh/koanf/v2"
	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
)

func TestWhenLoadingTokenSettings(t *testing.T) {
	t.Parallel()

	t.Run("Given a max tokens per request configuration", func(t *testing.T) {
		t.Parallel()

		k := koanf.New(".")
		_ = k.Load(confmap.Provider(map[string]any{
			"token.max_tokens_per_request": uint(1234),
		}, "."), nil)

		t.Run("Should load the max tokens per request as provided", func(t *testing.T) {
			t.Parallel()

			result := settings.NewTokenSettings(zap.NewNop(), k)

			assert.NotNil(t, result)
			assert.Equal(t, uint(1234), result.MaxTokensPerRequest)
		})
	})
}
