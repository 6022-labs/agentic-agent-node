package use_cases

import (
	"strings"

	"github.com/6022-labs/agentic-agent-node/src/agent_knowledge/models"
	"github.com/6022-labs/agentic-agent-node/src/agent_knowledge/repositories"
)

type AppendConversationMessages struct {
	repository repositories.ConversationMessageRepositoryInterface
}

func NewAppendConversationMessages(repository repositories.ConversationMessageRepositoryInterface) *AppendConversationMessages {
	return &AppendConversationMessages{repository: repository}
}

func (u *AppendConversationMessages) Execute(conversationId string, messages []models.ConversationMessage) ([]models.ConversationMessage, error) {
	conversationId = strings.TrimSpace(conversationId)
	if len(conversationId) == 0 || len(messages) == 0 {
		return nil, nil
	}

	prepared := make([]models.ConversationMessage, 0, len(messages))
	for _, msg := range messages {
		msg.ConversationId = conversationId
		prepared = append(prepared, msg)
	}
	return u.repository.AppendMessages(prepared)
}
