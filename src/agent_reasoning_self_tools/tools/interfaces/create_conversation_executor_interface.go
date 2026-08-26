package interfaces

import (
	"context"

	knowledge_requests "github.com/6022-labs/agentic-agent-node/src/agent_knowledge/use_cases/requests"
	knowledge_responses "github.com/6022-labs/agentic-agent-node/src/agent_knowledge/use_cases/responses"
)

// CreateConversationExecutorInterface exposes named memory-session creation to the self tools.
type CreateConversationExecutorInterface interface {
	Execute(
		ctx context.Context,
		request *knowledge_requests.CreateConversationRequest,
	) (*knowledge_responses.ConversationResponse, error)
}
