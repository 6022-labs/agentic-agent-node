package repositories

import (
	"context"

	"github.com/6022-labs/agentic-agent-node/src/agent_knowledge/models"
)

// ConversationMessageRepositoryInterface — dedup is the caller's job; repo
// keeps ON CONFLICT (hash) DO NOTHING as a race-safety net only.
type ConversationMessageRepositoryInterface interface {
	AppendMessages(ctx context.Context, messages []models.ConversationMessage) ([]models.ConversationMessage, error)
	ListByConversation(ctx context.Context, conversationId string) ([]models.ConversationMessage, error)
	ListConversationIds(ctx context.Context) ([]string, error)
	ListExistingHashes(ctx context.Context, hashes []string) ([]string, error)
}
