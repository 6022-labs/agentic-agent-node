package repositories

import (
	"context"

	"github.com/6022-labs/agentic-agent-node/src/agent_reasoning/models"
	"github.com/google/uuid"
)

// LLMConnectionRepositoryInterface persists an agent's LLM provider config.
// One connection per agent, so the agent holds zero or one row at a time.
type LLMConnectionRepositoryInterface interface {
	// GetByAgentId returns the agent's connection, or (nil, nil) if none has been set.
	GetByAgentId(ctx context.Context, agentId uuid.UUID) (*models.LLMConnection, error)
	// Upsert creates the row if absent, otherwise updates the agent's existing
	// row in place (preserving Id + CreatedAt). Scoped by connection.AgentId.
	Upsert(ctx context.Context, connection *models.LLMConnection) (*models.LLMConnection, error)
	// DeleteByAgentId clears the agent's connection. No-op if none is set.
	DeleteByAgentId(ctx context.Context, agentId uuid.UUID) error
}
