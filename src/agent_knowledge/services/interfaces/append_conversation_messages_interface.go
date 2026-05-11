package interfaces

import (
	"github.com/6022-labs/agentic-agent-node/src/agent_knowledge/models"
)

type AppendConversationMessagesInterface interface {
	Append(conversationId string, messages []models.ConversationMessage) ([]models.ConversationMessage, error)
}
