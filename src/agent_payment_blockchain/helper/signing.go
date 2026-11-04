package helper

import (
	"fmt"
	"math/big"
	"strings"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/crypto"
)

// Transactor builds a chain-bound transactor from a 0x-prefixed private key hex.
func Transactor(privateKeyHex string, chainId uint64) (*bind.TransactOpts, error) {
	hex := strings.TrimSpace(privateKeyHex)
	if len(hex) >= 2 && (hex[:2] == "0x" || hex[:2] == "0X") {
		hex = hex[2:]
	}
	key, err := crypto.HexToECDSA(hex)
	if err != nil {
		return nil, fmt.Errorf("invalid wallet private key: %w", err)
	}
	return bind.NewKeyedTransactorWithChainID(key, new(big.Int).SetUint64(chainId))
}
