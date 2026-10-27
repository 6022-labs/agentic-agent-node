package interfaces

import (
	"context"

	identity_requests "github.com/6022-labs/agentic-agent-node/src/agent_identity/use_cases/requests"
	identity_responses "github.com/6022-labs/agentic-agent-node/src/agent_identity/use_cases/responses"
)

// SignTypedDataExecutorInterface exposes raw EIP-712 signing to the self tools.
type SignTypedDataExecutorInterface interface {
	Execute(
		ctx context.Context,
		request *identity_requests.SignTypedDataRequest,
	) (*identity_responses.TypedDataSignatureResponse, error)
}
