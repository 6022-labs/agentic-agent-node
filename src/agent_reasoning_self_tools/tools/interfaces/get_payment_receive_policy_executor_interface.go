package interfaces

import (
	"context"

	payment_responses "github.com/6022-labs/agentic-agent-node/src/agent_payment/use_cases/responses"
)

// GetPaymentReceivePolicyExecutorInterface exposes the payment policy read to the self tools.
type GetPaymentReceivePolicyExecutorInterface interface {
	Execute(ctx context.Context) (*payment_responses.PaymentReceivePolicyResponse, error)
}
