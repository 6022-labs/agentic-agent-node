package interfaces

import (
	"context"

	"github.com/6022-labs/agentic-agent-node/src/gateway/models"
)

// PaymentPolicyClientInterface is the gateway-side port for the runtime's
// policy-head endpoints; Get materializes the default policy on first access.
type PaymentPolicyClientInterface interface {
	Get(ctx context.Context) (*models.ExternalPaymentPolicyPayload, error)
	Delete(ctx context.Context) error
}
