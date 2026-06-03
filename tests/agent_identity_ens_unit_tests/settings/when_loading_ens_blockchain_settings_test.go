package settings_test

import (
	"testing"

	ens_settings "github.com/6022-labs/agentic-agent-node/src/agent_identity_ens/settings"
	"github.com/knadh/koanf/providers/confmap"
	"github.com/knadh/koanf/v2"
	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
)

func TestWhenLoadingEnsBlockchainSettings(t *testing.T) {
	t.Parallel()

	t.Run("Given a configured chain id", func(t *testing.T) {
		t.Parallel()

		k := koanf.New(".")
		_ = k.Load(confmap.Provider(map[string]any{
			"blockchain.id": uint64(11155111),
		}, "."), nil)

		t.Run("Should load ChainId from the shared blockchain.id key", func(t *testing.T) {
			t.Parallel()

			result := ens_settings.NewEnsBlockchainSettings(zap.NewNop(), k)

			assert.NotNil(t, result)
			assert.Equal(t, uint64(11155111), result.ChainId)
		})
	})

	t.Run("Given unrelated blockchain config alongside the chain id", func(t *testing.T) {
		t.Parallel()

		k := koanf.New(".")
		_ = k.Load(confmap.Provider(map[string]any{
			"blockchain.id":           uint64(80002),
			"blockchain.rpc_http_url": "https://amoy.example.com",
		}, "."), nil)

		t.Run("Should read only the chain id and ignore the rest", func(t *testing.T) {
			t.Parallel()

			result := ens_settings.NewEnsBlockchainSettings(zap.NewNop(), k)

			assert.Equal(t, uint64(80002), result.ChainId)
		})
	})
}
