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

func TestWhenLoadingPaymentFacilitatorsSettings(t *testing.T) {
	t.Parallel()

	t.Run("Given configured facilitators", func(t *testing.T) {
		t.Parallel()

		k := koanf.New(".")
		_ = k.Load(confmap.Provider(map[string]any{
			"payment.facilitators": map[string]any{
				"eip155:80002": map[string]any{"base_url": "https://x402-amoy.polygon.technology/"},
				"eip155:84532": map[string]any{"base_url": "https://x402.org/facilitator"},
			},
		}, "."), nil)

		t.Run("Should resolve base urls and list the networks in order", func(t *testing.T) {
			t.Parallel()

			result := settings.NewPaymentFacilitatorsSettings(zap.NewNop(), k, reachableChains())

			amoy, ok := result.BaseUrlFor("eip155:80002")
			assert.True(t, ok)
			assert.Equal(t, "https://x402-amoy.polygon.technology", amoy)

			base, ok := result.BaseUrlFor("eip155:84532")
			assert.True(t, ok)
			assert.Equal(t, "https://x402.org/facilitator", base)

			assert.Equal(t, []string{"eip155:80002", "eip155:84532"}, result.Networks())
		})
	})

	t.Run("Given a facilitator with and without an explicit timeout", func(t *testing.T) {
		t.Parallel()

		k := koanf.New(".")
		_ = k.Load(confmap.Provider(map[string]any{
			"payment.facilitators": map[string]any{
				"eip155:80002": map[string]any{"base_url": "https://x402-amoy.polygon.technology", "max_timeout_seconds": 120},
				"eip155:84532": map[string]any{"base_url": "https://x402.org/facilitator"},
			},
		}, "."), nil)

		t.Run("Should keep the explicit timeout and default the missing one", func(t *testing.T) {
			t.Parallel()

			entries := settings.NewPaymentFacilitatorsSettings(zap.NewNop(), k, reachableChains()).Entries()

			if assert.Len(t, entries, 2) {
				assert.Equal(t, "eip155:80002", entries[0].Network)
				assert.Equal(t, 120, entries[0].MaxTimeoutSeconds)
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

			result := settings.NewPaymentFacilitatorsSettings(zap.NewNop(), k, reachableChains())

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
			"payment.facilitators": map[string]any{
				// 8453 has a facilitator but no entry in reachableChains().
				"eip155:8453": map[string]any{"base_url": "https://x402.org/facilitator"},
			},
		}, "."), nil)

		t.Run("Should fail fast — the node could never validate its assets", func(t *testing.T) {
			t.Parallel()

			// WriteThenPanic turns the constructor's logger.Fatal into a panic so
			// the boot-time guard is assertable without exiting the test process.
			logger := zap.NewNop().WithOptions(
				zap.WithFatalHook(zapcore.WriteThenPanic),
			)

			assert.Panics(t, func() {
				settings.NewPaymentFacilitatorsSettings(logger, k, reachableChains())
			})
		})
	})
}
