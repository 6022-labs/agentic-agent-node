package interfaces

import (
	"context"

	payment_requests "github.com/6022-labs/agentic-agent-node/src/agent_payment/use_cases/requests"
	payment_responses "github.com/6022-labs/agentic-agent-node/src/agent_payment/use_cases/responses"
)

// CreatePaymentReceiveRuleExecutorInterface exposes the payment rule create to the self tools.
type CreatePaymentReceiveRuleExecutorInterface interface {
	Execute(
		ctx context.Context,
		request *payment_requests.SavePaymentReceiveRuleRequest,
	) (*payment_responses.PaymentReceiveRuleResponse, error)
}
