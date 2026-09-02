package interfaces

import (
	"context"
	"math/big"
)

// X402PaymentSpendPolicyEvaluatorInterface gates one x402 quote against the agent's spend policy.
type X402PaymentSpendPolicyEvaluatorInterface interface {
	// IsAuthorized reports whether some x402 rule matches every facet of the quote and its cap covers the amount.
	IsAuthorized(ctx context.Context, quote *X402PaymentQuote, amount *big.Int) (bool, error)
}
