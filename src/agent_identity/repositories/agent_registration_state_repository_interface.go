package repositories

import (
	"context"

	"github.com/6022-labs/agentic-agent-node/src/agent_identity/models"
)

// AgentRegistrationStateRepositoryInterface persists the agent's mint + ENS state as a single row.
type AgentRegistrationStateRepositoryInterface interface {
	// Get returns the persisted state, or (nil, nil) if none has been set.
	Get(ctx context.Context) (*models.AgentRegistrationState, error)
	Upsert(ctx context.Context, state *models.AgentRegistrationState) (*models.AgentRegistrationState, error)
}
