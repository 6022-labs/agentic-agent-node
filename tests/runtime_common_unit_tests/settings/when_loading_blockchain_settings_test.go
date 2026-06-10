package settings_test

import (
	"testing"

	blockchain_settings "github.com/6022-labs/agentic-agent-node/src/runtime_common/settings"
	"github.com/knadh/koanf/providers/confmap"
	"github.com/knadh/koanf/v2"
	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
)

func TestWhenLoadingBlockchainSettings(t *testing.T) {
	t.Parallel()

	t.Run("Given a configured chain id", func(t *testing.T) {
		t.Parallel()

		k := koanf.New(".")
		_ = k.Load(confmap.Provider(map[string]any{
			"blockchain.id": uint64(11155111),
		}, "."), nil)

		t.Run("Should load the chain id", func(t *testing.T) {
			t.Parallel()

			result := blockchain_settings.NewBlockchainSettings(zap.NewNop(), k)

			assert.NotNil(t, result)
			assert.Equal(t, uint64(11155111), result.Id)
		})
	})

	t.Run("Given unrelated rpc config alongside the chain id", func(t *testing.T) {
		t.Parallel()

		k := koanf.New(".")
		_ = k.Load(confmap.Provider(map[string]any{
			"blockchain.id": uint64(80002),
			"rpc.http_url":  "https://amoy.example.com",
		}, "."), nil)

		t.Run("Should read only the chain id and ignore the rest", func(t *testing.T) {
			t.Parallel()

			result := blockchain_settings.NewBlockchainSettings(zap.NewNop(), k)

			assert.Equal(t, uint64(80002), result.Id)
		})
	})
}
