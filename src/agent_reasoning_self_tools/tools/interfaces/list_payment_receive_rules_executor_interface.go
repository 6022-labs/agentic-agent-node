package interfaces

import (
	"context"

	payment_responses "github.com/6022-labs/agentic-agent-node/src/agent_payment/use_cases/responses"
)

// ListPaymentReceiveRulesExecutorInterface exposes the payment rules list to the self tools.
type ListPaymentReceiveRulesExecutorInterface interface {
	Execute(ctx context.Context) (*payment_responses.ListPaymentReceiveRulesResponse, error)
}
