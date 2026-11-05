package settings_test

import (
	"testing"
	"time"

	"github.com/6022-labs/agentic-agent-node/src/common/host_configurations"
	"github.com/6022-labs/agentic-agent-node/src/gateway/core/settings"
	"github.com/knadh/koanf/providers/confmap"
	"github.com/knadh/koanf/v2"
	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
)

func TestWhenLoadingJwtSettings(t *testing.T) {
	t.Parallel()

	t.Run("Given a complete configuration", func(t *testing.T) {
		t.Parallel()

		k := koanf.New(".")
		_ = k.Load(confmap.Provider(map[string]any{
			"jwt.signing_key":      "test-signing-key-at-least-32-characters",
			"jwt.token_ttl":        900,
			"jwt.nonce_ttl":        120,
			"jwt.refresh_ttl":      86400,
			"jwt.refresh_idle_ttl": 3600,
		}, "."), nil)

		t.Run("Should load every field as provided", func(t *testing.T) {
			t.Parallel()

			result := settings.NewJwtSettings(zap.NewNop(), k)

			assert.NotNil(t, result)
			assert.Equal(t, "test-signing-key-at-least-32-characters", result.SigningKey)
			assert.Equal(t, 900*time.Second, result.TokenTtl)
			assert.Equal(t, 120*time.Second, result.NonceTtl)
			assert.Equal(t, 86400*time.Second, result.RefreshTtl)
			assert.Equal(t, time.Hour, result.RefreshIdleTtl)
		})
	})

	t.Run("Given only signing_key is provided", func(t *testing.T) {
		t.Parallel()

		k := koanf.New(".")
		_ = k.Load(confmap.Provider(map[string]any{
			"jwt.signing_key": "test-signing-key-at-least-32-characters",
		}, "."), nil)

		t.Run(
			"Should default token_ttl to 1800, nonce_ttl to 300, refresh_ttl to 7 days and refresh_idle_ttl to 1 day",
			func(t *testing.T) {
				t.Parallel()

				result := settings.NewJwtSettings(zap.NewNop(), k)

				assert.Equal(t, 1800*time.Second, result.TokenTtl)
				assert.Equal(t, 300*time.Second, result.NonceTtl)
				assert.Equal(t, 7*24*time.Hour, result.RefreshTtl)
				assert.Equal(t, 24*time.Hour, result.RefreshIdleTtl)
			},
		)
	})

	t.Run("Given a padded signing_key", func(t *testing.T) {
		t.Parallel()

		k := koanf.New(".")
		_ = k.Load(confmap.Provider(map[string]any{
			"jwt.signing_key": "  test-signing-key-at-least-32-characters  ",
		}, "."), nil)

		t.Run("Should trim signing_key", func(t *testing.T) {
			t.Parallel()

			result := settings.NewJwtSettings(zap.NewNop(), k)
			assert.Equal(t, "test-signing-key-at-least-32-characters", result.SigningKey)
		})
	})

	t.Run("Given non-positive ttls", func(t *testing.T) {
		t.Parallel()

		k := koanf.New(".")
		_ = k.Load(confmap.Provider(map[string]any{
			"jwt.signing_key":      "test-signing-key-at-least-32-characters",
			"jwt.token_ttl":        0,
			"jwt.nonce_ttl":        0,
			"jwt.refresh_ttl":      0,
			"jwt.refresh_idle_ttl": 0,
		}, "."), nil)

		t.Run("Should fall back to defaults", func(t *testing.T) {
			t.Parallel()

			result := settings.NewJwtSettings(zap.NewNop(), k)
			assert.Equal(t, 1800*time.Second, result.TokenTtl)
			assert.Equal(t, 300*time.Second, result.NonceTtl)
			assert.Equal(t, 7*24*time.Hour, result.RefreshTtl)
			assert.Equal(t, 24*time.Hour, result.RefreshIdleTtl)
		})
	})
}

func TestWhenLoadingJwtSettingsFromEnv(t *testing.T) {
	t.Run("Given an env override", func(t *testing.T) {
		t.Setenv("JWT__SIGNING_KEY", "test-signing-key-at-least-32-characters")
		t.Setenv("JWT__REFRESH_TTL", "3600")
		t.Setenv("JWT__REFRESH_IDLE_TTL", "600")

		k, err := host_configurations.LoadKoanfConfig()
		assert.NoError(t, err)

		t.Run("Should win over the defaults", func(t *testing.T) {
			result := settings.NewJwtSettings(zap.NewNop(), k)

			assert.Equal(t, time.Hour, result.RefreshTtl)
			assert.Equal(t, 10*time.Minute, result.RefreshIdleTtl)
		})
	})
}
