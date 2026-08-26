package interfaces

import (
	"context"

	"github.com/6022-labs/agentic-agent-node/src/agent_reasoning/models"
)

// ConversationTitlerInterface labels a conversation from its opening message.
type ConversationTitlerInterface interface {
	// TitleIfUnnamed is a no-op when the conversation already carries a name.
	TitleIfUnnamed(
		ctx context.Context,
		conversationId string,
		llmConnection *models.LLMConnection,
		firstUserMessage string,
	) error
}
