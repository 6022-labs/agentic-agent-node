// Package helper holds blockchain primitives shared by every on-chain adapter.
package helper

import (
	"context"
	"fmt"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/ethclient"
)

// WaitForReceipt blocks until tx is mined and errors if it reverted.
func WaitForReceipt(ctx context.Context, client *ethclient.Client, tx *types.Transaction) (*types.Receipt, error) {
	receipt, err := bind.WaitMined(ctx, client, tx)
	if err != nil {
		return nil, fmt.Errorf("wait for tx %s: %w", tx.Hash().Hex(), err)
	}
	if receipt.Status != types.ReceiptStatusSuccessful {
		return nil, fmt.Errorf("tx %s reverted (status=%d)", tx.Hash().Hex(), receipt.Status)
	}
	return receipt, nil
}

// WaitForSuccess discards the receipt; use WaitForReceipt when the logs are needed.
func WaitForSuccess(ctx context.Context, client *ethclient.Client, tx *types.Transaction) error {
	_, err := WaitForReceipt(ctx, client, tx)
	return err
}
