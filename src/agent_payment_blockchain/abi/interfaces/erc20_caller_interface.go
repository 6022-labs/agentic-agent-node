package interfaces

import (
	"context"
	"math/big"

	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/common"
)

// Erc20CallerInterface re-declares bind.ContractCaller so the mock tooling
// can see it; *ethclient.Client implements it.
type Erc20CallerInterface interface {
	CodeAt(ctx context.Context, contract common.Address, blockNumber *big.Int) ([]byte, error)
	CallContract(ctx context.Context, msg ethereum.CallMsg, blockNumber *big.Int) ([]byte, error)
}
