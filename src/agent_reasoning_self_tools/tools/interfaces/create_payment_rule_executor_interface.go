package interfaces

import (
	"context"
	payment_requests "github.com/6022-labs/agentic-agent-node/src/agent_payment/use_cases/requests"
	payment_responses "github.com/6022-labs/agentic-agent-node/src/agent_payment/use_cases/responses"
)

// CreatePaymentRuleExecutorInterface exposes the payment rule create to the self tools.
type CreatePaymentRuleExecutorInterface interface {
	Execute(ctx context.Context, request *payment_requests.SavePaymentRuleRequest) (*payment_responses.PaymentRuleResponse, error)
}
