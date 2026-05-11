package interfaces

import (
	"github.com/6022-labs/agentic-agent-node/src/agent_knowledge/models"
)

type ListConversationMessagesInterface interface {
	List(conversationId string) ([]models.ConversationMessage, error)
}
