package settings_test

import (
	"testing"

	"github.com/6022-labs/agentic-agent-node/src/runtime_common/settings"
	"github.com/ethereum/go-ethereum/common"
	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
)

func TestWhenLoadingAgentCollectionCreatorsSettings(t *testing.T) {
	t.Parallel()

	t.Run("Given a configured chain with a built-in creator", func(t *testing.T) {
		t.Parallel()

		chains := settings.EvmChainsSettings{
			"80002": {ChainId: "80002", HttpUrl: "http://amoy"},
		}

		t.Run("Should resolve a creator address", func(t *testing.T) {
			t.Parallel()

			result := settings.NewAgentCollectionCreatorsSettings(zap.NewNop(), &chains)

			addr, ok := result.AddressFor("80002")
			assert.True(t, ok)
			assert.NotEqual(t, common.Address{}, addr)
		})
	})

	t.Run("Given a configured chain whose built-in creator entry is empty", func(t *testing.T) {
		t.Parallel()

		chains := settings.EvmChainsSettings{
			"137": {ChainId: "137", HttpUrl: "http://polygon"},
		}

		t.Run("Should skip it", func(t *testing.T) {
			t.Parallel()

			result := settings.NewAgentCollectionCreatorsSettings(zap.NewNop(), &chains)

			_, ok := result.AddressFor("137")
			assert.False(t, ok)
		})
	})

	t.Run("Given a built-in creator on an unconfigured chain", func(t *testing.T) {
		t.Parallel()

		chains := settings.EvmChainsSettings{
			"84532": {ChainId: "84532", HttpUrl: "http://base-sepolia"},
		}

		t.Run("Should not expose it", func(t *testing.T) {
			t.Parallel()

			result := settings.NewAgentCollectionCreatorsSettings(zap.NewNop(), &chains)

			_, ok := result.AddressFor("80002")
			assert.False(t, ok)
		})
	})
}
