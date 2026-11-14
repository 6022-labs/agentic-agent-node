package settings_test

import (
	"testing"

	"github.com/6022-labs/agentic-agent-node/src/agent_identity/settings"
	"github.com/6022-labs/agentic-agent-node/src/common/host_configurations"
	"github.com/knadh/koanf/providers/confmap"
	"github.com/knadh/koanf/v2"
	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
	"time"
)

func TestWhenLoadingAgentIdentitySettings(t *testing.T) {
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

	t.Run("Given no identity sync interval is configured", func(t *testing.T) {
		t.Parallel()

		k := koanf.New(".")
		_ = k.Load(confmap.Provider(map[string]any{
			"agent_identity.runtime_url": "https://agent.example.com",
		}, "."), nil)

		t.Run("Should default IdentitySyncInterval to 10 minutes", func(t *testing.T) {
			t.Parallel()

			result := settings.NewAgentIdentitySettings(zap.NewNop(), k)

			assert.Equal(t, 10*time.Minute, result.IdentitySyncInterval)
		})
	})

	t.Run("Given a custom identity sync interval", func(t *testing.T) {
		t.Parallel()

		k := koanf.New(".")
		_ = k.Load(confmap.Provider(map[string]any{
			"agent_identity.runtime_url":                    "https://agent.example.com",
			"agent_identity.identity_sync_interval_seconds": 90,
		}, "."), nil)

		t.Run("Should convert it to IdentitySyncInterval", func(t *testing.T) {
			t.Parallel()

			result := settings.NewAgentIdentitySettings(zap.NewNop(), k)

			assert.Equal(t, 90*time.Second, result.IdentitySyncInterval)
		})
	})

	t.Run("Given an env override", func(t *testing.T) {
		t.Setenv("AGENT_IDENTITY__RUNTIME_URL", "https://agent.example.com")
		t.Setenv("AGENT_IDENTITY__IDENTITY_SYNC_INTERVAL_SECONDS", "45")

		k, err := host_configurations.LoadKoanfConfig()
		assert.NoError(t, err)

		t.Run("Should win over the defaults", func(t *testing.T) {
			result := settings.NewAgentIdentitySettings(zap.NewNop(), k)

			assert.Equal(t, 45*time.Second, result.IdentitySyncInterval)
		})
	})
}
