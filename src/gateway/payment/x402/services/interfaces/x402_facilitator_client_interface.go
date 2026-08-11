package interfaces

import (
	"context"

	"github.com/6022-labs/agentic-agent-node/src/gateway/payment/x402"
)

// X402FacilitatorClientInterface is the port for an x402 v2 facilitator;
// baseUrl comes from the matched rule (one facilitator per network).
type X402FacilitatorClientInterface interface {
	// Verify asks the facilitator whether the payment is valid and fundable.
	Verify(
		ctx context.Context,
		baseUrl string,
		payload *x402.PaymentPayload,
		requirements *x402.PaymentRequirements,
	) (*x402.VerifyResponse, error)
	// Settle submits the payment on-chain and reports the transaction.
	Settle(
		ctx context.Context,
		baseUrl string,
		payload *x402.PaymentPayload,
		requirements *x402.PaymentRequirements,
	) (*x402.SettleResponse, error)
}
