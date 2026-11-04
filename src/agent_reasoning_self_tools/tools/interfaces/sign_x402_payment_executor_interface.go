package interfaces

import (
	"context"

	payment_requests "github.com/6022-labs/agentic-agent-node/src/agent_payment/use_cases/requests"
	payment_responses "github.com/6022-labs/agentic-agent-node/src/agent_payment/use_cases/responses"
)

// SignX402PaymentExecutorInterface exposes x402 payment signing to the self tools.
type SignX402PaymentExecutorInterface interface {
	Execute(
		ctx context.Context,
		request *payment_requests.SignX402PaymentRequest,
	) (*payment_responses.X402PaymentSignatureResponse, error)
}
