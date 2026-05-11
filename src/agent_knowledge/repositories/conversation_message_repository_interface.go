package repositories

import "github.com/6022-labs/agentic-agent-node/src/agent_knowledge/models"

// ConversationMessageRepositoryInterface — dedup is the caller's job; repo
// keeps ON CONFLICT (hash) DO NOTHING as a race-safety net only.
type ConversationMessageRepositoryInterface interface {
	AppendMessages(messages []models.ConversationMessage) ([]models.ConversationMessage, error)
	ListByConversation(conversationId string) ([]models.ConversationMessage, error)
	ListExistingHashes(hashes []string) ([]string, error)
}
