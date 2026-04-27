package repositories

import "github.com/6022-labs/agentic-agent-node/src/agent_knowledge/models"

// ConversationMessageRepositoryInterface persists conversation messages locally
// and lets use cases query them. Implementations must enforce uniqueness on
// `Hash` so repeated AppendMessages calls with semantically identical inputs
// are idempotent.
type ConversationMessageRepositoryInterface interface {
	// AppendMessages upserts the slice; existing hashes are skipped silently.
	// Returns the messages that were actually inserted.
	AppendMessages(messages []models.ConversationMessage) ([]models.ConversationMessage, error)
	// ListByConversation returns every message for a conversation, ordered by
	// timestamp ascending.
	ListByConversation(conversationId string) ([]models.ConversationMessage, error)
}
