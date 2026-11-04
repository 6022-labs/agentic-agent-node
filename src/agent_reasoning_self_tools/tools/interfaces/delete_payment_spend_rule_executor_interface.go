package interfaces

import (
	"context"

	payment_requests "github.com/6022-labs/agentic-agent-node/src/agent_payment/use_cases/requests"
)

// DeletePaymentSpendRuleExecutorInterface exposes the spend rule delete to the self tools.
type DeletePaymentSpendRuleExecutorInterface interface {
	Execute(ctx context.Context, request *payment_requests.DeletePaymentSpendRuleRequest) error
}
