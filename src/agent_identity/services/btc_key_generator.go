package services

import (
	"encoding/hex"

	"github.com/6022-labs/agentic-agent-node/src/agent_identity/models/enums"
	identity_interfaces "github.com/6022-labs/agentic-agent-node/src/agent_identity/services/interfaces"
	runtime_common_settings "github.com/6022-labs/agentic-agent-node/src/runtime_common/settings"
	"github.com/btcsuite/btcd/btcec/v2"
	"github.com/btcsuite/btcd/btcutil"
)

// BtcKeyGenerator generates a secp256k1 keypair and a native-segwit (P2WPKH) address for the node's bitcoin network.
type BtcKeyGenerator struct {
	btcChainsSettings *runtime_common_settings.BtcChainsSettings
}

func NewBtcKeyGenerator(btcChainsSettings *runtime_common_settings.BtcChainsSettings) *BtcKeyGenerator {
	return &BtcKeyGenerator{btcChainsSettings: btcChainsSettings}
}

func (g *BtcKeyGenerator) Type() enums.AddressType {
	return enums.AddressTypeBtc
}

func (g *BtcKeyGenerator) Generate() (*identity_interfaces.WalletCredentials, error) {
	key, err := btcec.NewPrivateKey()
	if err != nil {
		return nil, err
	}
	witnessProgram := btcutil.Hash160(key.PubKey().SerializeCompressed())
	addr, err := btcutil.NewAddressWitnessPubKeyHash(witnessProgram, g.btcChainsSettings.ChainParams())
	if err != nil {
		return nil, err
	}
	return &identity_interfaces.WalletCredentials{
		PrivateKeyHex: hex.EncodeToString(key.Serialize()),
		Address:       addr.EncodeAddress(),
	}, nil
}
