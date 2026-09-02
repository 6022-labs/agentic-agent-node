package interfaces

import "github.com/6022-labs/agentic-agent-node/src/agent_payment/models/enums"

// PaymentReceiveMethodStrategiesInterface resolves the strategy for a payment method
// (the Method discriminator). Use cases depend on this so they can be tested
// with a mock registry.
type PaymentReceiveMethodStrategiesInterface interface {
	// For returns the strategy for the method, or (nil, false) if unsupported.
	For(method enums.PaymentMethod) (PaymentReceiveMethodStrategyInterface, bool)
}
