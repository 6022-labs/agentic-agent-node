package interfaces

import (
	"context"

	payment_responses "github.com/6022-labs/agentic-agent-node/src/agent_payment/use_cases/responses"
)

// GetPaymentSpendPolicyExecutorInterface exposes the spend policy head to the self tools.
type GetPaymentSpendPolicyExecutorInterface interface {
	Execute(ctx context.Context) (*payment_responses.PaymentSpendPolicyResponse, error)
}
