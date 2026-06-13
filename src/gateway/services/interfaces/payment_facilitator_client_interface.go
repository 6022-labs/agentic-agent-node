package interfaces

import "context"

// PaymentFacilitatorClientInterface is the gateway-side port for the runtime's
// facilitator config.
type PaymentFacilitatorClientInterface interface {
	// Get returns the network → facilitator base URL mapping the gate settles
	// through.
	Get(ctx context.Context) (map[string]string, error)
}
