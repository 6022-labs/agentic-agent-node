package interfaces

import (
	"context"
	"math/big"
)

// X402PaymentSpendPolicyEvaluatorInterface gates one x402 quote against the agent's spend policy.
type X402PaymentSpendPolicyEvaluatorInterface interface {
	// Authorize reports whether a matching rule covers the amount both per payment and within
	// its recurring budget. It is not a predicate: a true verdict has recorded the spend.
	Authorize(ctx context.Context, quote *X402PaymentQuote, amount *big.Int) (bool, error)
}
