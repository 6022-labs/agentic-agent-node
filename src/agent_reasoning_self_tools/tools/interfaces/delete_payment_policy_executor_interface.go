package interfaces

import (
	"context"
)

// DeletePaymentPolicyExecutorInterface exposes the payment policy delete to the self tools.
type DeletePaymentPolicyExecutorInterface interface {
	Execute(ctx context.Context) error
}
