package settings_test

import (
	"testing"

	"github.com/6022-labs/agentic-agent-node/src/runtime_common/settings"
	"github.com/knadh/koanf/providers/confmap"
	"github.com/knadh/koanf/v2"
	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

func TestWhenLoadingPaymentReceiveSettings(t *testing.T) {
	t.Parallel()

	t.Run("Given configured facilitators", func(t *testing.T) {
		t.Parallel()

		k := koanf.New(".")
		_ = k.Load(confmap.Provider(map[string]any{
			"payment.receive.facilitators": map[string]any{
				"80002": map[string]any{"base_url": "https://x402-amoy.polygon.technology/"},
				"84532": map[string]any{"base_url": "https://x402.org/facilitator"},
			},
		}, "."), nil)

		t.Run("Should resolve the base urls", func(t *testing.T) {
			t.Parallel()

			result := settings.NewPaymentReceiveSettings(zap.NewNop(), k, reachableChains())

			amoy, ok := result.BaseUrlFor("eip155:80002")
			assert.True(t, ok)
			assert.Equal(t, "https://x402-amoy.polygon.technology", amoy)

			base, ok := result.BaseUrlFor("eip155:84532")
			assert.True(t, ok)
			assert.Equal(t, "https://x402.org/facilitator", base)
		})

		t.Run("Should list the networks in order", func(t *testing.T) {
			t.Parallel()

			result := settings.NewPaymentReceiveSettings(zap.NewNop(), k, reachableChains())

			assert.Equal(t, []string{"eip155:80002", "eip155:84532"}, result.Networks())
		})
	})

	t.Run("Given a facilitator with and without an explicit timeout", func(t *testing.T) {
		t.Parallel()

		k := koanf.New(".")
		_ = k.Load(confmap.Provider(map[string]any{
			"payment.receive.facilitators": map[string]any{
				"80002": map[string]any{"base_url": "https://x402-amoy.polygon.technology", "max_timeout_seconds": 120},
				"84532": map[string]any{"base_url": "https://x402.org/facilitator"},
			},
		}, "."), nil)

		t.Run("Should keep the explicit timeout", func(t *testing.T) {
			t.Parallel()

			entries := settings.NewPaymentReceiveSettings(zap.NewNop(), k, reachableChains()).Entries()

			if assert.Len(t, entries, 2) {
				assert.Equal(t, "eip155:80002", entries[0].Network)
				assert.Equal(t, 120, entries[0].MaxTimeoutSeconds)
			}
		})

		t.Run("Should default the missing timeout", func(t *testing.T) {
			t.Parallel()

			entries := settings.NewPaymentReceiveSettings(zap.NewNop(), k, reachableChains()).Entries()

			if assert.Len(t, entries, 2) {
				assert.Equal(t, "eip155:84532", entries[1].Network)
				assert.Equal(t, 60, entries[1].MaxTimeoutSeconds)
			}
		})
	})

	t.Run("Given a network with no configured facilitator", func(t *testing.T) {
		t.Parallel()

		k := koanf.New(".")

		t.Run("Should report it as not configured", func(t *testing.T) {
			t.Parallel()

			result := settings.NewPaymentReceiveSettings(zap.NewNop(), k, reachableChains())

			baseUrl, ok := result.BaseUrlFor("eip155:8453")
			assert.False(t, ok)
			assert.Empty(t, baseUrl)
			assert.Empty(t, result.Networks())
		})
	})

	t.Run("Given a facilitator whose chain has no RPC", func(t *testing.T) {
		t.Parallel()

		k := koanf.New(".")
		_ = k.Load(confmap.Provider(map[string]any{
			"payment.receive.facilitators": map[string]any{
				// 8453 has a facilitator but no entry in reachableChains().
				"8453": map[string]any{"base_url": "https://x402.org/facilitator"},
			},
		}, "."), nil)

		t.Run("Should fail fast — the node could never validate its assets", func(t *testing.T) {
			t.Parallel()

			// WriteThenPanic turns logger.Fatal into a panic so the guard is assertable.
			logger := zap.NewNop().WithOptions(
				zap.WithFatalHook(zapcore.WriteThenPanic),
			)

			assert.Panics(t, func() {
				settings.NewPaymentReceiveSettings(logger, k, reachableChains())
			})
		})
	})

	t.Run("Given the legacy payment.facilitators key", func(t *testing.T) {
		t.Parallel()

		k := koanf.New(".")
		_ = k.Load(confmap.Provider(map[string]any{
			"payment.facilitators": map[string]any{
				"80002": map[string]any{"base_url": "https://x402-amoy.polygon.technology"},
			},
		}, "."), nil)

		t.Run("Should refuse to boot rather than settle nothing", func(t *testing.T) {
			t.Parallel()

			fatalLogger := zap.NewNop().WithOptions(zap.WithFatalHook(zapcore.WriteThenPanic))

			assert.Panics(t, func() {
				settings.NewPaymentReceiveSettings(fatalLogger, k, reachableChains())
			})
		})
	})
}
