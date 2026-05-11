package services

import (
	"strings"

	"github.com/6022-labs/agentic-agent-node/src/agent_knowledge/models"
	"github.com/6022-labs/agentic-agent-node/src/agent_knowledge/repositories"
)

type ListConversationMessages struct {
	repository repositories.ConversationMessageRepositoryInterface
}

func NewListConversationMessages(repository repositories.ConversationMessageRepositoryInterface) *ListConversationMessages {
	return &ListConversationMessages{repository: repository}
}

func (s *ListConversationMessages) List(conversationId string) ([]models.ConversationMessage, error) {
	conversationId = strings.TrimSpace(conversationId)
	if len(conversationId) == 0 {
		return nil, nil
	}
	return s.repository.ListByConversation(conversationId)
}
