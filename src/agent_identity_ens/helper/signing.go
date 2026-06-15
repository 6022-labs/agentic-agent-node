package helper

import (
	"context"
	"fmt"
	"math/big"
	"strings"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethclient"
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

// WaitForSuccess blocks until tx is mined and errors if it reverted.
func WaitForSuccess(ctx context.Context, client *ethclient.Client, tx *types.Transaction) error {
	receipt, err := bind.WaitMined(ctx, client, tx)
	if err != nil {
		return fmt.Errorf("wait for tx %s: %w", tx.Hash().Hex(), err)
	}
	if receipt.Status != types.ReceiptStatusSuccessful {
		return fmt.Errorf("tx %s reverted (status=%d)", tx.Hash().Hex(), receipt.Status)
	}
	return nil
}
