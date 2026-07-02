package settings_test

import (
	"testing"

	"github.com/6022-labs/agentic-agent-node/src/runtime_common/settings"
	"github.com/knadh/koanf/providers/confmap"
	"github.com/knadh/koanf/v2"
	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
)

func TestWhenLoadingEvmChainsSettings(t *testing.T) {
	t.Parallel()

	t.Run("Given configured chains", func(t *testing.T) {
		t.Parallel()

		k := koanf.New(".")
		_ = k.Load(confmap.Provider(map[string]any{
			"blockchain.evm_chains": map[string]any{
				"80002": map[string]any{"http_url": "https://rpc-amoy.polygon.technology"},
				"84532": map[string]any{"http_url": "https://sepolia.base.org"},
			},
		}, "."), nil)

		t.Run("Should resolve http urls and list the chain ids in order", func(t *testing.T) {
			t.Parallel()

			result := settings.NewEvmChainsSettings(zap.NewNop(), k)

			assert.True(t, result.Has("80002"))
			assert.True(t, result.Has("84532"))
			assert.False(t, result.Has("1"))

			amoy, ok := result.HttpUrlFor("80002")
			assert.True(t, ok)
			assert.Equal(t, "https://rpc-amoy.polygon.technology", amoy)

			base, ok := result.HttpUrlFor("84532")
			assert.True(t, ok)
			assert.Equal(t, "https://sepolia.base.org", base)

			assert.Equal(t, []string{"80002", "84532"}, result.ChainIds())
		})
	})

	t.Run("Given an unconfigured chain", func(t *testing.T) {
		t.Parallel()

		k := koanf.New(".")
		_ = k.Load(confmap.Provider(map[string]any{
			"blockchain.evm_chains": map[string]any{
				"80002": map[string]any{"http_url": "https://rpc-amoy.polygon.technology"},
			},
		}, "."), nil)

		t.Run("Should report it as not reachable", func(t *testing.T) {
			t.Parallel()

			result := settings.NewEvmChainsSettings(zap.NewNop(), k)

			httpUrl, ok := result.HttpUrlFor("1")
			assert.False(t, ok)
			assert.Empty(t, httpUrl)
		})
	})
}
