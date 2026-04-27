package settings_test

import (
	"testing"

	blockchain_settings "github.com/6022-labs/agentic-agent-node/src/agent_identity_blockchain/settings"
	"github.com/knadh/koanf/providers/confmap"
	"github.com/knadh/koanf/v2"
	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
)

func TestWhenLoadingBlockchainSettings(t *testing.T) {
	t.Parallel()

	t.Run("Given a complete configuration", func(t *testing.T) {
		t.Parallel()

		k := koanf.New(".")
		_ = k.Load(confmap.Provider(map[string]any{
			"blockchain.id":           uint64(11155111),
			"blockchain.rpc_http_url": "https://sepolia.example.com",
		}, "."), nil)

		t.Run("Should load every field as provided", func(t *testing.T) {
			t.Parallel()

			result := blockchain_settings.NewBlockchainSettings(zap.NewNop(), k)

			assert.NotNil(t, result)
			assert.Equal(t, uint64(11155111), result.Id)
			assert.Equal(t, "https://sepolia.example.com", result.RpcHttpUrl)
		})
	})

	t.Run("Given rpc_http_url has surrounding whitespace", func(t *testing.T) {
		t.Parallel()

		k := koanf.New(".")
		_ = k.Load(confmap.Provider(map[string]any{
			"blockchain.id":           uint64(1),
			"blockchain.rpc_http_url": "  https://eth.example.com  ",
		}, "."), nil)

		t.Run("Should trim it", func(t *testing.T) {
			t.Parallel()

			result := blockchain_settings.NewBlockchainSettings(zap.NewNop(), k)
			assert.Equal(t, "https://eth.example.com", result.RpcHttpUrl)
		})
	})
}
