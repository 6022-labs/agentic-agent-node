package interfaces

import (
	"context"

	"github.com/6022-labs/agentic-agent-node/src/agent_knowledge/models"
)

type ListConversationMessagesInterface interface {
	List(ctx context.Context, conversationId string) ([]models.ConversationMessage, error)
}
