package interfaces

import (
	"context"

	"github.com/6022-labs/agentic-agent-node/src/gateway/payment/models"
)

// PaymentReceiveAccessClientInterface is the gateway-side port for the runtime's access decision.
type PaymentReceiveAccessClientInterface interface {
	// Authorize sends whichever credential the caller presented; both may be empty.
	Authorize(
		ctx context.Context,
		address string,
		paymentSignature string,
	) (*models.ExternalPaymentReceiveAccessDecision, error)
}
