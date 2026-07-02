package services_test

import (
	"strings"
	"testing"

	"github.com/6022-labs/agentic-agent-node/src/agent_identity/models/enums"
	identity_services "github.com/6022-labs/agentic-agent-node/src/agent_identity/services"
	runtime_common_settings "github.com/6022-labs/agentic-agent-node/src/runtime_common/settings"
	"github.com/knadh/koanf/v2"
	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
)

func newBtcSettings(t *testing.T, network string) *runtime_common_settings.BtcChainsSettings {
	t.Helper()
	k := koanf.New(".")
	_ = k.Set(runtime_common_settings.BtcNetworkSettingsKey, network)
	return runtime_common_settings.NewBtcChainsSettings(zap.NewNop(), k)
}

func TestWhenGeneratingBtcKey(t *testing.T) {
	t.Parallel()

	t.Run("Given the btc key generator", func(t *testing.T) {
		t.Parallel()

		t.Run("Should report the btc address type", func(t *testing.T) {
			t.Parallel()

			sut := identity_services.NewBtcKeyGenerator(newBtcSettings(t, "testnet"))

			assert.Equal(t, enums.AddressTypeBtc, sut.Type())
		})

		t.Run("Should generate a testnet native-segwit address", func(t *testing.T) {
			t.Parallel()

			sut := identity_services.NewBtcKeyGenerator(newBtcSettings(t, "testnet"))

			cred, err := sut.Generate()

			assert.NoError(t, err)
			assert.NotEmpty(t, cred.PrivateKeyHex)
			assert.True(t, strings.HasPrefix(cred.Address, "tb1"))
		})

		t.Run("Should generate a mainnet native-segwit address", func(t *testing.T) {
			t.Parallel()

			sut := identity_services.NewBtcKeyGenerator(newBtcSettings(t, "mainnet"))

			cred, err := sut.Generate()

			assert.NoError(t, err)
			assert.NotEmpty(t, cred.PrivateKeyHex)
			assert.True(t, strings.HasPrefix(cred.Address, "bc1"))
		})
	})
}
