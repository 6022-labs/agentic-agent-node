package services

import (
	"strings"

	"github.com/6022-labs/agentic-agent-node/src/agent_knowledge/models"
	"github.com/6022-labs/agentic-agent-node/src/agent_knowledge/repositories"
	runtime_utils "github.com/6022-labs/agentic-agent-node/src/runtime_common/utils"
)

// AppendConversationMessages demojizes content, stamps conversationId,
// hashes, drops dupes against stored rows, then inserts the rest.
type AppendConversationMessages struct {
	repository repositories.ConversationMessageRepositoryInterface
}

func NewAppendConversationMessages(repository repositories.ConversationMessageRepositoryInterface) *AppendConversationMessages {
	return &AppendConversationMessages{repository: repository}
}

func (s *AppendConversationMessages) Append(conversationId string, messages []models.ConversationMessage) ([]models.ConversationMessage, error) {
	conversationId = strings.TrimSpace(conversationId)
	if len(conversationId) == 0 || len(messages) == 0 {
		return nil, nil
	}

	prepared := make([]models.ConversationMessage, 0, len(messages))
	hashes := make([]string, 0, len(messages))
	for _, msg := range messages {
		msg.ConversationId = conversationId
		msg.Content = runtime_utils.DemojizeContent(msg.Content)

		if err := msg.ComputeHash(); err != nil {
			return nil, err
		}

		prepared = append(prepared, msg)
		hashes = append(hashes, msg.Hash)
	}

	existing, err := s.repository.ListExistingHashes(hashes)
	if err != nil {
		return nil, err
	}
	existingSet := make(map[string]struct{}, len(existing))
	for _, h := range existing {
		existingSet[h] = struct{}{}
	}

	toInsert := make([]models.ConversationMessage, 0, len(prepared))
	for _, msg := range prepared {
		if _, hit := existingSet[msg.Hash]; hit {
			continue
		}
		toInsert = append(toInsert, msg)
	}
	if len(toInsert) == 0 {
		return nil, nil
	}

	return s.repository.AppendMessages(toInsert)
}
