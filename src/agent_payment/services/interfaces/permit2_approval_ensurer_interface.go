package interfaces

import (
	"context"
	"math/big"
)

// Permit2ApprovalEnsurerInterface guarantees the node wallet has approved the
// Permit2 contract, without which a signed permit2 authorization can never settle.
type Permit2ApprovalEnsurerInterface interface {
	// EnsureApproval sends approve(Permit2, max) when the allowance is short of
	// amount and blocks until it lands; network is CAIP-2 (eip155:80002).
	EnsureApproval(ctx context.Context, network string, asset string, amount *big.Int) error
}
