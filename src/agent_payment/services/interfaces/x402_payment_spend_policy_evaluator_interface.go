package interfaces

import (
	"context"
	"math/big"
	"time"

	"github.com/google/uuid"
)

// X402PaymentSpendReservation is what Authorize recorded, so a payment that is never signed can give it back.
type X402PaymentSpendReservation struct {
	RuleId uuid.UUID
	Amount *big.Int
	// WindowStarts maps each charged window to the period the amount was charged in.
	WindowStarts map[uuid.UUID]time.Time
}

// X402PaymentSpendPolicyEvaluatorInterface gates one x402 quote against the agent's spend policy.
type X402PaymentSpendPolicyEvaluatorInterface interface {
	// Authorize records the spend and returns its reservation, or nil when the policy refuses the quote.
	// It is not a predicate: call Release if the payment is then not signed.
	Authorize(ctx context.Context, quote *X402PaymentQuote, amount *big.Int) (*X402PaymentSpendReservation, error)
	// Release gives a reservation back; a period that rolled over since keeps the spend it now holds.
	Release(ctx context.Context, reservation *X402PaymentSpendReservation) error
}
