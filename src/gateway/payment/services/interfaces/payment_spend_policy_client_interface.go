package interfaces

import (
	"context"

	"github.com/6022-labs/agentic-agent-node/src/gateway/payment/models"
)

// PaymentSpendPolicyClientInterface is the gateway-side port for the runtime's spend policy head.
type PaymentSpendPolicyClientInterface interface {
	Get(ctx context.Context) (*models.ExternalPaymentSpendPolicyPayload, error)
	Update(
		ctx context.Context,
		payload *models.ExternalUpdatePaymentSpendPolicyPayload,
	) (*models.ExternalPaymentSpendPolicyPayload, error)
	Delete(ctx context.Context) error
}
