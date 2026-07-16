package interfaces

import (
	"context"

	payment_requests "github.com/6022-labs/agentic-agent-node/src/agent_payment/use_cases/requests"
)

// DeletePaymentRuleExecutorInterface exposes the payment rule delete to the self tools.
type DeletePaymentRuleExecutorInterface interface {
	Execute(ctx context.Context, request *payment_requests.DeletePaymentRuleRequest) error
}
