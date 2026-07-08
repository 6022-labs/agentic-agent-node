package settings_test

import (
	"testing"
	"time"

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
			"jwt.signing_key": "test-signing-key-at-least-32-characters",
			"jwt.token_ttl":   900,
			"jwt.nonce_ttl":   120,
		}, "."), nil)

		t.Run("Should load every field as provided", func(t *testing.T) {
			t.Parallel()

			result := settings.NewJwtSettings(zap.NewNop(), k)

			assert.NotNil(t, result)
			assert.Equal(t, "test-signing-key-at-least-32-characters", result.SigningKey)
			assert.Equal(t, 900*time.Second, result.TokenTtl)
			assert.Equal(t, 120*time.Second, result.NonceTtl)
		})
	})

	t.Run("Given only signing_key is provided", func(t *testing.T) {
		t.Parallel()

		k := koanf.New(".")
		_ = k.Load(confmap.Provider(map[string]any{
			"jwt.signing_key": "test-signing-key-at-least-32-characters",
		}, "."), nil)

		t.Run("Should default token_ttl to 1800 and nonce_ttl to 300", func(t *testing.T) {
			t.Parallel()

			result := settings.NewJwtSettings(zap.NewNop(), k)

			assert.Equal(t, 1800*time.Second, result.TokenTtl)
			assert.Equal(t, 300*time.Second, result.NonceTtl)
		})
	})

	t.Run("Given signing_key has surrounding whitespace", func(t *testing.T) {
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
			"jwt.signing_key": "test-signing-key-at-least-32-characters",
			"jwt.token_ttl":   0,
			"jwt.nonce_ttl":   0,
		}, "."), nil)

		t.Run("Should fall back to defaults", func(t *testing.T) {
			t.Parallel()

			result := settings.NewJwtSettings(zap.NewNop(), k)
			assert.Equal(t, 1800*time.Second, result.TokenTtl)
			assert.Equal(t, 300*time.Second, result.NonceTtl)
		})
	})
}
