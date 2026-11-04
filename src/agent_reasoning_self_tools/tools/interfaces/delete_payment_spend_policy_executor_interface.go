package interfaces

import "context"

// DeletePaymentSpendPolicyExecutorInterface exposes the spend policy reset to the self tools.
type DeletePaymentSpendPolicyExecutorInterface interface {
	Execute(ctx context.Context) error
}
