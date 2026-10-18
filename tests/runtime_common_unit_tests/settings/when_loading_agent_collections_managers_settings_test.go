package settings_test

import (
	"testing"

	"github.com/6022-labs/agentic-agent-node/src/runtime_common/settings"
	"github.com/ethereum/go-ethereum/common"
	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
)

func TestWhenLoadingAgentCollectionsManagersSettings(t *testing.T) {
	t.Parallel()

	t.Run("Given a configured chain with a built-in manager", func(t *testing.T) {
		t.Parallel()

		chains := settings.EvmChainsSettings{
			"80002": {ChainId: "80002", HttpUrl: "http://amoy"},
		}

		t.Run("Should resolve a manager address", func(t *testing.T) {
			t.Parallel()

			result := settings.NewAgentCollectionsManagersSettings(zap.NewNop(), &chains)

			addr, ok := result.AddressFor("80002")
			assert.True(t, ok)
			assert.NotEqual(t, common.Address{}, addr)
		})

		t.Run("Should list the chain", func(t *testing.T) {
			t.Parallel()

			result := settings.NewAgentCollectionsManagersSettings(zap.NewNop(), &chains)

			assert.Equal(t, []string{"80002"}, result.ChainIds())
		})
	})

	t.Run("Given a built-in manager on an unconfigured chain", func(t *testing.T) {
		t.Parallel()

		chains := settings.EvmChainsSettings{
			"80002": {ChainId: "80002", HttpUrl: "http://amoy"},
		}

		t.Run("Should not expose it", func(t *testing.T) {
			t.Parallel()

			result := settings.NewAgentCollectionsManagersSettings(zap.NewNop(), &chains)

			addr, ok := result.AddressFor("137")
			assert.False(t, ok)
			assert.Equal(t, common.Address{}, addr)
		})
	})

	t.Run("Given a chain without a built-in manager", func(t *testing.T) {
		t.Parallel()

		chains := settings.EvmChainsSettings{
			"84532": {ChainId: "84532", HttpUrl: "http://base-sepolia"},
		}

		t.Run("Should report it as not available", func(t *testing.T) {
			t.Parallel()

			result := settings.NewAgentCollectionsManagersSettings(zap.NewNop(), &chains)

			_, ok := result.AddressFor("84532")
			assert.False(t, ok)
			assert.Empty(t, result.ChainIds())
		})
	})
}
