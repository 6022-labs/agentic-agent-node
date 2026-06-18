package interfaces

import (
	"math/big"

	"github.com/ethereum/go-ethereum/common"
)

// Token6022RequesterInterface reads the $6022 ERC-20 to rate an agent by its creator's holdings.
type Token6022RequesterInterface interface {
	BalanceOf(chainId string, account common.Address) (*big.Int, error)
	// Decimals is the token's decimal precision.
	Decimals() (uint8, error)
}
