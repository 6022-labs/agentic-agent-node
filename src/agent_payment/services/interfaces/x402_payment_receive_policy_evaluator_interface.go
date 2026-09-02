package interfaces

import (
	"context"

	"github.com/6022-labs/agentic-agent-node/src/agent_payment/models/enums"
	"github.com/6022-labs/agentic-agent-node/src/common/x402"
)

// PaymentReceiveAccessCredentials is whatever the caller presented; both fields may be empty.
type PaymentReceiveAccessCredentials struct {
	// Address is a caller the gateway already proved (JWT); it buys free access only.
	Address          string
	PaymentSignature string
}

// PaymentReceiveAccessDecision is the verdict plus what executing it needs.
type PaymentReceiveAccessDecision struct {
	Outcome    enums.PaymentReceiveAccessOutcome
	Payer      string
	ErrorCode  string
	Accepts    []x402.PaymentRequirements
	Settlement *PaymentReceiveSettlement
}

// PaymentReceiveSettlement pairs the matched requirements with the facilitator that can settle them.
type PaymentReceiveSettlement struct {
	Requirements       x402.PaymentRequirements
	FacilitatorBaseUrl string
	Network            string
}

// X402PaymentReceivePolicyEvaluatorInterface decides whether a caller may reach a gated route.
type X402PaymentReceivePolicyEvaluatorInterface interface {
	// Evaluate returns the verdict plus what the gateway needs to advertise a 402 or settle a payment.
	Evaluate(ctx context.Context, credentials *PaymentReceiveAccessCredentials) (*PaymentReceiveAccessDecision, error)
}
