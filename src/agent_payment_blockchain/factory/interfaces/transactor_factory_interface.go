package interfaces

import (
	"context"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
)

// TransactorFactoryInterface builds a chain-bound transactor signed by the node wallet.
type TransactorFactoryInterface interface {
	Create(ctx context.Context, chainId uint64) (*bind.TransactOpts, error)
}
