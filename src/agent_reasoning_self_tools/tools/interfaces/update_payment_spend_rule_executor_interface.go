package interfaces

import (
	"context"

	payment_requests "github.com/6022-labs/agentic-agent-node/src/agent_payment/use_cases/requests"
	payment_responses "github.com/6022-labs/agentic-agent-node/src/agent_payment/use_cases/responses"
)

// UpdatePaymentSpendRuleExecutorInterface exposes the spend rule update to the self tools.
type UpdatePaymentSpendRuleExecutorInterface interface {
	Execute(
		ctx context.Context,
		request *payment_requests.UpdatePaymentSpendRuleRequest,
	) (*payment_responses.PaymentSpendRuleResponse, error)
}
