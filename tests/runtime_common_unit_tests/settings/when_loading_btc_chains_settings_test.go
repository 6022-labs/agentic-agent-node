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

func TestWhenLoadingBtcChainsSettings(t *testing.T) {
	t.Parallel()

	t.Run("Given no configured network", func(t *testing.T) {
		t.Parallel()

		t.Run("Should default to mainnet", func(t *testing.T) {
			t.Parallel()

			result := settings.NewBtcChainsSettings(zap.NewNop(), koanf.New("."))

			assert.Equal(t, "bitcoin:mainnet", result.NetworkId())
			assert.Equal(t, "mainnet", result.ChainParams().Name)
		})
	})

	t.Run("Given testnet is configured", func(t *testing.T) {
		t.Parallel()

		k := koanf.New(".")
		_ = k.Load(confmap.Provider(map[string]any{"blockchain.btc_chains.network": "testnet"}, "."), nil)

		t.Run("Should resolve the testnet id and params", func(t *testing.T) {
			t.Parallel()

			result := settings.NewBtcChainsSettings(zap.NewNop(), k)

			assert.Equal(t, "bitcoin:testnet", result.NetworkId())
			assert.Equal(t, "testnet3", result.ChainParams().Name)
		})
	})

	t.Run("Given an invalid network", func(t *testing.T) {
		t.Parallel()

		k := koanf.New(".")
		_ = k.Load(confmap.Provider(map[string]any{"blockchain.btc_chains.network": "regtest"}, "."), nil)

		t.Run("Should fail fast", func(t *testing.T) {
			t.Parallel()

			logger := zap.NewNop().WithOptions(zap.WithFatalHook(zapcore.WriteThenPanic))

			assert.Panics(t, func() {
				settings.NewBtcChainsSettings(logger, k)
			})
		})
	})
}
