package interfaces

import (
	"context"

	"github.com/6022-labs/agentic-agent-node/src/gateway/models"
)

// PaymentFacilitatorClientInterface is the gateway-side port for the runtime's
// facilitator config.
type PaymentFacilitatorClientInterface interface {
	// Get returns the network → facilitator mapping (base URL + settlement
	// timeout) the gate settles through.
	Get(ctx context.Context) (map[string]models.ExternalPaymentFacilitator, error)
}
