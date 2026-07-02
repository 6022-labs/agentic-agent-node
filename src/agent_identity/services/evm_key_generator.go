package services

import (
	"github.com/6022-labs/agentic-agent-node/src/agent_identity/models/enums"
	identity_interfaces "github.com/6022-labs/agentic-agent-node/src/agent_identity/services/interfaces"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
)

// EvmKeyGenerator generates a secp256k1 keypair and its 0x EVM address.
type EvmKeyGenerator struct{}

func NewEvmKeyGenerator() *EvmKeyGenerator {
	return &EvmKeyGenerator{}
}

func (g *EvmKeyGenerator) Type() enums.AddressType {
	return enums.AddressTypeEvm
}

func (g *EvmKeyGenerator) Generate() (*identity_interfaces.WalletCredentials, error) {
	key, err := crypto.GenerateKey()
	if err != nil {
		return nil, err
	}
	pkHex := "0x" + common.Bytes2Hex(crypto.FromECDSA(key))
	addr := crypto.PubkeyToAddress(key.PublicKey).Hex()
	return &identity_interfaces.WalletCredentials{PrivateKeyHex: pkHex, Address: addr}, nil
}
