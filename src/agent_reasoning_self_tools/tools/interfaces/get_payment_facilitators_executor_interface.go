package interfaces

import (
	"context"
	payment_responses "github.com/6022-labs/agentic-agent-node/src/agent_payment/use_cases/responses"
)

// GetPaymentFacilitatorsExecutorInterface exposes the facilitators read to the self tools.
type GetPaymentFacilitatorsExecutorInterface interface {
	Execute(ctx context.Context) (*payment_responses.PaymentFacilitatorsResponse, error)
}
