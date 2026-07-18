package settings_test

import (
	"testing"

	"github.com/6022-labs/agentic-agent-node/src/runtime_common/settings"
	"github.com/ethereum/go-ethereum/common"
	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
)

func TestWhenLoadingToken6022Settings(t *testing.T) {
	t.Parallel()

	t.Run("Given a configured chain with a built-in token", func(t *testing.T) {
		t.Parallel()

		chains := settings.EvmChainsSettings{
			"80002": {ChainId: "80002", HttpUrl: "http://amoy"},
		}

		t.Run("Should resolve a token address", func(t *testing.T) {
			t.Parallel()

			result := settings.NewToken6022Settings(zap.NewNop(), &chains)

			addr, ok := result.AddressFor("80002")
			assert.True(t, ok)
			assert.NotEqual(t, common.Address{}, addr)
		})
	})

	t.Run("Given a built-in token on an unconfigured chain", func(t *testing.T) {
		t.Parallel()

		chains := settings.EvmChainsSettings{
			"80002": {ChainId: "80002", HttpUrl: "http://amoy"},
		}

		t.Run("Should not expose it", func(t *testing.T) {
			t.Parallel()

			result := settings.NewToken6022Settings(zap.NewNop(), &chains)

			addr, ok := result.AddressFor("137")
			assert.False(t, ok)
			assert.Equal(t, common.Address{}, addr)
		})
	})

	t.Run("Given a chain without a built-in token", func(t *testing.T) {
		t.Parallel()

		chains := settings.EvmChainsSettings{
			"31337": {ChainId: "31337", HttpUrl: "http://localhost:8545"},
		}

		t.Run("Should report it as not available", func(t *testing.T) {
			t.Parallel()

			result := settings.NewToken6022Settings(zap.NewNop(), &chains)

			_, ok := result.AddressFor("31337")
			assert.False(t, ok)
		})
	})
}
