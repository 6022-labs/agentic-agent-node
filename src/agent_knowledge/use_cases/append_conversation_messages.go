package use_cases

import (
	"github.com/6022-labs/agentic-agent-node/src/agent_knowledge/models"
	"github.com/6022-labs/agentic-agent-node/src/agent_knowledge/services/interfaces"
)

type AppendConversationMessagesInterface interface {
	Execute(conversationId string, messages []models.ConversationMessage) ([]models.ConversationMessage, error)
}

type AppendConversationMessages struct {
	service interfaces.AppendConversationMessagesInterface
}

func NewAppendConversationMessages(service interfaces.AppendConversationMessagesInterface) *AppendConversationMessages {
	return &AppendConversationMessages{service: service}
}

func (u *AppendConversationMessages) Execute(conversationId string, messages []models.ConversationMessage) ([]models.ConversationMessage, error) {
	return u.service.Append(conversationId, messages)
}
