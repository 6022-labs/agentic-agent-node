package settings_test

import (
	"testing"

	"github.com/6022-labs/agentic-agent-node/src/agent_identity/settings"
	"github.com/knadh/koanf/providers/confmap"
	"github.com/knadh/koanf/v2"
	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
)

func TestWhenLoadingAgentIdentitySettings(t *testing.T) {
	t.Parallel()

	t.Run("Given a configured runtime url", func(t *testing.T) {
		t.Parallel()

		k := koanf.New(".")
		_ = k.Load(confmap.Provider(map[string]any{
			"agent_identity.runtime_url": "https://agent.example.com",
		}, "."), nil)

		t.Run("Should load RuntimeUrl from the agent_identity.runtime_url key", func(t *testing.T) {
			t.Parallel()

			result := settings.NewAgentIdentitySettings(zap.NewNop(), k)

			assert.NotNil(t, result)
			assert.Equal(t, "https://agent.example.com", result.RuntimeUrl)
		})
	})

	t.Run("Given a padded runtime url", func(t *testing.T) {
		t.Parallel()

		k := koanf.New(".")
		_ = k.Load(confmap.Provider(map[string]any{
			"agent_identity.runtime_url": "  https://agent.example.com  ",
		}, "."), nil)

		t.Run("Should trim the surrounding whitespace", func(t *testing.T) {
			t.Parallel()

			result := settings.NewAgentIdentitySettings(zap.NewNop(), k)

			assert.Equal(t, "https://agent.example.com", result.RuntimeUrl)
		})
	})

	t.Run("Given no ens base domain is configured", func(t *testing.T) {
		t.Parallel()

		k := koanf.New(".")
		_ = k.Load(confmap.Provider(map[string]any{
			"agent_identity.runtime_url": "https://agent.example.com",
		}, "."), nil)

		t.Run("Should default EnsBaseDomain to 6022.eth", func(t *testing.T) {
			t.Parallel()

			result := settings.NewAgentIdentitySettings(zap.NewNop(), k)

			assert.Equal(t, settings.DefaultEnsBaseDomain, result.EnsBaseDomain)
			assert.Equal(t, "6022.eth", result.EnsBaseDomain)
		})
	})

	t.Run("Given a padded custom ens base domain", func(t *testing.T) {
		t.Parallel()

		k := koanf.New(".")
		_ = k.Load(confmap.Provider(map[string]any{
			"agent_identity.runtime_url":     "https://agent.example.com",
			"agent_identity.ens_base_domain": "  agent.6022.eth  ",
		}, "."), nil)

		t.Run("Should load and trim it", func(t *testing.T) {
			t.Parallel()

			result := settings.NewAgentIdentitySettings(zap.NewNop(), k)

			assert.Equal(t, "agent.6022.eth", result.EnsBaseDomain)
		})
	})
}
