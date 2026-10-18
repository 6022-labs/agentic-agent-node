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
	// GetById returns (nil, nil) when absent; ownership checks are the caller's job.
	GetById(ctx context.Context, id string) (*models.Conversation, error)
	// GetByIdAndAgentIdAndCallerId returns (nil, nil) when absent or out of scope; nil callerId is unscoped.
	GetByIdAndAgentIdAndCallerId(
		ctx context.Context,
		id string,
		agentId uuid.UUID,
		callerId *string,
	) (*models.Conversation, error)
	// ExistsByAgentIdAndCallerIdAndName: nil callerId means caller-less heads; excludedId skips the renamed one.
	ExistsByAgentIdAndCallerIdAndName(
		ctx context.Context,
		agentId uuid.UUID,
		callerId *string,
		name string,
		excludedId *string,
	) (bool, error)
	// Update persists the whole head.
	Update(ctx context.Context, conversation *models.Conversation) (*models.Conversation, error)
	// ListByAgentId lists newest first; nil callerId, offset or limit means unscoped/unbounded.
	ListByAgentId(
		ctx context.Context,
		agentId uuid.UUID,
		callerId *string,
		offset *int,
		limit *int,
	) ([]models.Conversation, error)
	// CountByAgentId counts heads; nil callerId means unscoped.
	CountByAgentId(ctx context.Context, agentId uuid.UUID, callerId *string) (uint, error)
}
