package interfaces

import (
	"context"

	knowledge_requests "github.com/6022-labs/agentic-agent-node/src/agent_knowledge/use_cases/requests"
	knowledge_responses "github.com/6022-labs/agentic-agent-node/src/agent_knowledge/use_cases/responses"
)

// ListConversationsExecutorInterface exposes the paginated memory-session list to the self tools.
type ListConversationsExecutorInterface interface {
	Execute(
		ctx context.Context,
		request *knowledge_requests.ListConversationsRequest,
	) (*knowledge_responses.ListConversationsResponse, error)
}
