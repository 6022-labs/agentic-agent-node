package repositories

import (
	"context"

	"github.com/6022-labs/agentic-agent-node/src/agent_identity/models"
	"github.com/google/uuid"
)

// AgentRegistrationStateRepositoryInterface persists an agent's mint + ENS state as a single row per agent.
type AgentRegistrationStateRepositoryInterface interface {
	// GetByAgentId returns the agent's persisted state, or (nil, nil) if none has been set.
	GetByAgentId(ctx context.Context, agentId uuid.UUID) (*models.AgentRegistrationState, error)
	Upsert(ctx context.Context, state *models.AgentRegistrationState) (*models.AgentRegistrationState, error)
}
