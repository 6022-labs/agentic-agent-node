package interfaces

import (
	"context"

	knowledge_interfaces "github.com/6022-labs/agentic-agent-node/src/agent_knowledge/services/interfaces"
)

type ResponseDriverInterface interface {
	// Drive runs the inference loop under a freshly built prompt (instructions feed its
	// per-turn overlay; neither persists), then persists new inputs + final assistant (nil if none).
	Drive(
		ctx context.Context,
		conversationId string,
		instructions string,
		inputs []knowledge_interfaces.ConversationMessage,
	) (*knowledge_interfaces.ConversationMessage, error)
}
