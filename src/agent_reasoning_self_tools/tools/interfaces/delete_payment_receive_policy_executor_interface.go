package interfaces

import (
	"context"
)

// DeletePaymentReceivePolicyExecutorInterface exposes the payment policy delete to the self tools.
type DeletePaymentReceivePolicyExecutorInterface interface {
	Execute(ctx context.Context) error
}
