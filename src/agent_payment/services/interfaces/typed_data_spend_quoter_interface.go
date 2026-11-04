package interfaces

import (
	"math/big"

	"github.com/ethereum/go-ethereum/signer/core/apitypes"
)

// TypedDataSpendQuoterInterface reads what an EIP-712 payload would let someone take from the wallet.
type TypedDataSpendQuoterInterface interface {
	// Quote returns nil when the payload moves no funds, and refuses a token payload it cannot bound.
	Quote(typedData *apitypes.TypedData) (*X402PaymentQuote, *big.Int, error)
}
