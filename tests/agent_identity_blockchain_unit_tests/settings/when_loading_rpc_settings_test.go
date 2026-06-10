package settings_test

import (
	"testing"

	rpc_settings "github.com/6022-labs/agentic-agent-node/src/agent_identity_blockchain/settings"
	"github.com/knadh/koanf/providers/confmap"
	"github.com/knadh/koanf/v2"
	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
)

func TestWhenLoadingRpcSettings(t *testing.T) {
	t.Parallel()

	t.Run("Given a configured rpc endpoint", func(t *testing.T) {
		t.Parallel()

		k := koanf.New(".")
		_ = k.Load(confmap.Provider(map[string]any{
			"rpc.http_url": "https://sepolia.example.com",
		}, "."), nil)

		t.Run("Should load the url as provided", func(t *testing.T) {
			t.Parallel()

			result := rpc_settings.NewRpcSettings(zap.NewNop(), k)

			assert.NotNil(t, result)
			assert.Equal(t, "https://sepolia.example.com", result.HttpUrl)
		})
	})

	t.Run("Given the url has surrounding whitespace", func(t *testing.T) {
		t.Parallel()

		k := koanf.New(".")
		_ = k.Load(confmap.Provider(map[string]any{
			"rpc.http_url": "  https://eth.example.com  ",
		}, "."), nil)

		t.Run("Should trim it", func(t *testing.T) {
			t.Parallel()

			result := rpc_settings.NewRpcSettings(zap.NewNop(), k)

			assert.Equal(t, "https://eth.example.com", result.HttpUrl)
		})
	})
}
