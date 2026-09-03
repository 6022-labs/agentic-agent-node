package interfaces

import (
	"context"

	payment_responses "github.com/6022-labs/agentic-agent-node/src/agent_payment/use_cases/responses"
)

// ListPaymentSpendRulesExecutorInterface exposes the spend rules read to the self tools.
type ListPaymentSpendRulesExecutorInterface interface {
	Execute(ctx context.Context) (*payment_responses.ListPaymentSpendRulesResponse, error)
}
