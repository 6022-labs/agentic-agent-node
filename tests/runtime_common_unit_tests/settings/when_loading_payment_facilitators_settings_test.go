package settings_test

import (
	"testing"

	"github.com/6022-labs/agentic-agent-node/src/runtime_common/settings"
	"github.com/knadh/koanf/providers/confmap"
	"github.com/knadh/koanf/v2"
	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
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

			result := settings.NewPaymentFacilitatorsSettings(zap.NewNop(), k)

			amoy, ok := result.BaseUrlFor("eip155:80002")
			assert.True(t, ok)
			assert.Equal(t, "https://x402-amoy.polygon.technology", amoy)

			base, ok := result.BaseUrlFor("eip155:84532")
			assert.True(t, ok)
			assert.Equal(t, "https://x402.org/facilitator", base)

			assert.Equal(t, []string{"eip155:80002", "eip155:84532"}, result.Networks())
		})
	})

	t.Run("Given a network with no configured facilitator", func(t *testing.T) {
		t.Parallel()

		k := koanf.New(".")

		t.Run("Should report it as not configured", func(t *testing.T) {
			t.Parallel()

			result := settings.NewPaymentFacilitatorsSettings(zap.NewNop(), k)

			baseUrl, ok := result.BaseUrlFor("eip155:8453")
			assert.False(t, ok)
			assert.Empty(t, baseUrl)
			assert.Empty(t, result.Networks())
		})
	})
}
