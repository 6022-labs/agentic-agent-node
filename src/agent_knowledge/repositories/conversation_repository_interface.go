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
	// Create inserts a new head; the caller owns id generation and name uniqueness.
	Create(ctx context.Context, conversation *models.Conversation) (*models.Conversation, error)
	// GetById returns the conversation by its (globally unique) id, or (nil, nil)
	// if absent. Ownership is the caller's to enforce via the row's AgentId.
	GetById(ctx context.Context, id string) (*models.Conversation, error)
	// ExistsByAgentIdAndName reports whether the agent already uses that name; excludedId lets a rename ignore
	// the conversation being renamed.
	ExistsByAgentIdAndName(ctx context.Context, agentId uuid.UUID, name string, excludedId *string) (bool, error)
	// Update persists the whole head.
	Update(ctx context.Context, conversation *models.Conversation) (*models.Conversation, error)
	// ListByAgentId returns the agent's conversations, newest first; a nil offset or limit means that bound is
	// not applied.
	ListByAgentId(ctx context.Context, agentId uuid.UUID, offset *int, limit *int) ([]models.Conversation, error)
	// CountByAgentId totals the agent's heads for pagination.
	CountByAgentId(ctx context.Context, agentId uuid.UUID) (uint, error)
}
