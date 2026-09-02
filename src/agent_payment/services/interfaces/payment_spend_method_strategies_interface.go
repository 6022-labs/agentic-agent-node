package interfaces

import "github.com/6022-labs/agentic-agent-node/src/agent_payment/models/enums"

// PaymentSpendMethodStrategiesInterface resolves the spend strategy for a payment method.
type PaymentSpendMethodStrategiesInterface interface {
	// For returns the strategy for the method, or (nil, false) if unsupported.
	For(method enums.PaymentMethod) (PaymentSpendMethodStrategyInterface, bool)
}
