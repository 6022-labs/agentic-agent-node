package interfaces

import (
	"context"

	payment_requests "github.com/6022-labs/agentic-agent-node/src/agent_payment/use_cases/requests"
	payment_responses "github.com/6022-labs/agentic-agent-node/src/agent_payment/use_cases/responses"
)

// UpdatePaymentRuleExecutorInterface exposes the payment rule update to the self tools.
type UpdatePaymentRuleExecutorInterface interface {
	Execute(ctx context.Context, request *payment_requests.UpdatePaymentRuleRequest) (*payment_responses.PaymentRuleResponse, error)
}
