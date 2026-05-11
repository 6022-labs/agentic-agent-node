package use_cases

import (
	"github.com/6022-labs/agentic-agent-node/src/agent_knowledge/models"
	"github.com/6022-labs/agentic-agent-node/src/agent_knowledge/services/interfaces"
)

type ListConversationMessagesInterface interface {
	Execute(conversationId string) ([]models.ConversationMessage, error)
}

type ListConversationMessages struct {
	service interfaces.ListConversationMessagesInterface
}

func NewListConversationMessages(service interfaces.ListConversationMessagesInterface) *ListConversationMessages {
	return &ListConversationMessages{service: service}
}

func (u *ListConversationMessages) Execute(conversationId string) ([]models.ConversationMessage, error) {
	return u.service.List(conversationId)
}
