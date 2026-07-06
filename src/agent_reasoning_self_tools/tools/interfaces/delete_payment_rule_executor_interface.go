package interfaces

import (
	"context"
	"github.com/google/uuid"
)

// DeletePaymentRuleExecutorInterface exposes the payment rule delete to the self tools.
type DeletePaymentRuleExecutorInterface interface {
	Execute(ctx context.Context, id uuid.UUID) error
}
