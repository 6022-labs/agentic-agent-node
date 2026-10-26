package interfaces

import (
	"context"

	"github.com/6022-labs/agentic-agent-node/src/gateway/payment/models"
)

// PaymentReceivePolicyClientInterface is the gateway-side port for the runtime's
// policy-head endpoints; Get materializes the default policy on first access.
type PaymentReceivePolicyClientInterface interface {
	Get(ctx context.Context) (*models.ExternalPaymentReceivePolicyPayload, error)
	Delete(ctx context.Context) error
}
