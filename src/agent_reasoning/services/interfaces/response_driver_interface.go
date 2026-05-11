package interfaces

import (
	"context"

	knowledge_interfaces "github.com/6022-labs/agentic-agent-node/src/agent_knowledge/services/interfaces"
)

type ResponseDriverInterface interface {
	// Drive runs the inference loop and persists new inputs + final assistant
	// after the loop. Returns the persisted assistant row, or nil if the loop
	// produced no assistant text.
	Drive(
		ctx context.Context,
		conversationId string,
		inputs []knowledge_interfaces.ConversationMessage,
	) (*knowledge_interfaces.ConversationMessage, error)
}
