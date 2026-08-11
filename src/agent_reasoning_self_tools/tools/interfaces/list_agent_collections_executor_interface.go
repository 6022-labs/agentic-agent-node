package interfaces

import (
	"context"

	identity_requests "github.com/6022-labs/agentic-agent-node/src/agent_identity/use_cases/requests"
	identity_responses "github.com/6022-labs/agentic-agent-node/src/agent_identity/use_cases/responses"
)

// ListAgentCollectionsExecutorInterface exposes the collections read to the self tools.
type ListAgentCollectionsExecutorInterface interface {
	Execute(
		ctx context.Context,
		req *identity_requests.ListAgentCollectionsRequest,
	) (*identity_responses.AgentCollectionsResponse, error)
}
