package interfaces

import (
	"context"

	payment_requests "github.com/6022-labs/agentic-agent-node/src/agent_payment/use_cases/requests"
	payment_responses "github.com/6022-labs/agentic-agent-node/src/agent_payment/use_cases/responses"
)

// UpdatePaymentSpendPolicyExecutorInterface exposes the payee allowlist update to the self tools.
type UpdatePaymentSpendPolicyExecutorInterface interface {
	Execute(
		ctx context.Context,
		request *payment_requests.UpdatePaymentSpendPolicyRequest,
	) (*payment_responses.PaymentSpendPolicyResponse, error)
}
