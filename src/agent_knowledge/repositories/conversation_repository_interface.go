package repositories

import (
	"context"

	"github.com/6022-labs/agentic-agent-node/src/agent_knowledge/models"
	"github.com/google/uuid"
)

// ConversationRepositoryInterface persists the agent-scoped conversation heads
// that ConversationMessage rows hang off of.
type ConversationRepositoryInterface interface {
	// Ensure creates the head for (AgentId, Id) if absent; no-op if it exists.
	// Returns true when a new head was created.
	Ensure(ctx context.Context, conversation *models.Conversation) (bool, error)
	// GetById returns the conversation by its (globally unique) id, or (nil, nil)
	// if absent. Ownership is the caller's to enforce via the row's AgentId.
	GetById(ctx context.Context, id string) (*models.Conversation, error)
	// ListByAgentId returns the agent's conversations.
	ListByAgentId(ctx context.Context, agentId uuid.UUID) ([]models.Conversation, error)
}
