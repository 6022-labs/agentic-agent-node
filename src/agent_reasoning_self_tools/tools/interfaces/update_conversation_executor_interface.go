package interfaces

import (
	"context"

	knowledge_requests "github.com/6022-labs/agentic-agent-node/src/agent_knowledge/use_cases/requests"
	knowledge_responses "github.com/6022-labs/agentic-agent-node/src/agent_knowledge/use_cases/responses"
)

// UpdateConversationExecutorInterface exposes memory-session renaming to the self tools.
type UpdateConversationExecutorInterface interface {
	Execute(
		ctx context.Context,
		request *knowledge_requests.UpdateConversationRequest,
	) (*knowledge_responses.ConversationResponse, error)
}
