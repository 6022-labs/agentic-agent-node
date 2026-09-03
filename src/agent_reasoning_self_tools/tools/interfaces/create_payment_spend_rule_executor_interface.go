package interfaces

import (
	"context"

	payment_requests "github.com/6022-labs/agentic-agent-node/src/agent_payment/use_cases/requests"
	payment_responses "github.com/6022-labs/agentic-agent-node/src/agent_payment/use_cases/responses"
)

// CreatePaymentSpendRuleExecutorInterface exposes the spend rule create to the self tools.
type CreatePaymentSpendRuleExecutorInterface interface {
	Execute(
		ctx context.Context,
		request *payment_requests.SavePaymentSpendRuleRequest,
	) (*payment_responses.PaymentSpendRuleResponse, error)
}
