package settings_test

import (
	"testing"

	"github.com/6022-labs/agentic-agent-node/src/gateway/identity/settings"
	"github.com/knadh/koanf/providers/confmap"
	"github.com/knadh/koanf/v2"
	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
)

func TestWhenLoadingAgentIdentitySettings(t *testing.T) {
	t.Parallel()

	t.Run("Given runtime_url with a trailing slash", func(t *testing.T) {
		t.Parallel()

		k := koanf.New(".")
		_ = k.Load(confmap.Provider(map[string]any{
			"agent_identity.runtime_url": "https://agent.example.com/",
		}, "."), nil)

		t.Run("Should trim the trailing slash", func(t *testing.T) {
			t.Parallel()

			result := settings.NewAgentIdentitySettings(zap.NewNop(), k)

			assert.NotNil(t, result)
			assert.Equal(t, "https://agent.example.com", result.RuntimeUrl)
		})
	})

	t.Run("Given runtime_url with surrounding whitespace", func(t *testing.T) {
		t.Parallel()

		k := koanf.New(".")
		_ = k.Load(confmap.Provider(map[string]any{
			"agent_identity.runtime_url": "  https://agent.example.com  ",
		}, "."), nil)

		t.Run("Should trim the whitespace", func(t *testing.T) {
			t.Parallel()

			result := settings.NewAgentIdentitySettings(zap.NewNop(), k)
			assert.Equal(t, "https://agent.example.com", result.RuntimeUrl)
		})
	})
}
