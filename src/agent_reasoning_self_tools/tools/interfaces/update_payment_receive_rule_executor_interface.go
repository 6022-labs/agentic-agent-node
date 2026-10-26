package interfaces

import (
	"context"

	payment_requests "github.com/6022-labs/agentic-agent-node/src/agent_payment/use_cases/requests"
	payment_responses "github.com/6022-labs/agentic-agent-node/src/agent_payment/use_cases/responses"
)

// UpdatePaymentReceiveRuleExecutorInterface exposes the payment rule update to the self tools.
type UpdatePaymentReceiveRuleExecutorInterface interface {
	Execute(
		ctx context.Context,
		request *payment_requests.UpdatePaymentReceiveRuleRequest,
	) (*payment_responses.PaymentReceiveRuleResponse, error)
}
