package interfaces

import (
	"context"

	payment_requests "github.com/6022-labs/agentic-agent-node/src/agent_payment/use_cases/requests"
)

// DeletePaymentReceiveRuleExecutorInterface exposes the payment rule delete to the self tools.
type DeletePaymentReceiveRuleExecutorInterface interface {
	Execute(ctx context.Context, request *payment_requests.DeletePaymentReceiveRuleRequest) error
}
