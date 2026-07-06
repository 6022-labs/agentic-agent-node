package interfaces

import (
	"context"
	payment_requests "github.com/6022-labs/agentic-agent-node/src/agent_payment/use_cases/requests"
	payment_responses "github.com/6022-labs/agentic-agent-node/src/agent_payment/use_cases/responses"
	"github.com/google/uuid"
)

// UpdatePaymentRuleExecutorInterface exposes the payment rule update to the self tools.
type UpdatePaymentRuleExecutorInterface interface {
	Execute(ctx context.Context, id uuid.UUID, request *payment_requests.SavePaymentRuleRequest) (*payment_responses.PaymentRuleResponse, error)
}
