package interfaces

import (
	"context"

	payment_requests "github.com/6022-labs/agentic-agent-node/src/agent_payment/use_cases/requests"
	payment_responses "github.com/6022-labs/agentic-agent-node/src/agent_payment/use_cases/responses"
)

// SignTypedDataExecutorInterface exposes raw EIP-712 signing to the self tools.
type SignTypedDataExecutorInterface interface {
	Execute(
		ctx context.Context,
		request *payment_requests.SignTypedDataRequest,
	) (*payment_responses.TypedDataSignatureResponse, error)
}
