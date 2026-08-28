package repositories

import (
	"context"

	"github.com/6022-labs/agentic-agent-node/src/agent_knowledge/models"
)

// ConversationMessageRepositoryInterface — Manage messages for a conversation
type ConversationMessageRepositoryInterface interface {
	AppendMessages(ctx context.Context, messages []models.ConversationMessage) ([]models.ConversationMessage, error)
	ListByConversationId(ctx context.Context, conversationId string) ([]models.ConversationMessage, error)
	ListExistingHashes(ctx context.Context, hashes []string) ([]string, error)
	// CountByConversationIds returns the message count per conversation; ids with no message are absent.
	CountByConversationIds(ctx context.Context, conversationIds []string) (map[string]int64, error)
}
