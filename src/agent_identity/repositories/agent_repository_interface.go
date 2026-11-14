package repositories

import (
	"context"

	"github.com/6022-labs/agentic-agent-node/src/agent_identity/models"
)

// AgentRepositoryInterface persists the agent's identity as a single row.
type AgentRepositoryInterface interface {
	// Get returns the persisted agent, or (nil, nil) if none has been set.
	Get(ctx context.Context) (*models.Agent, error)
	Upsert(ctx context.Context, agent *models.Agent) (*models.Agent, error)
	// UpdateOnChainIdentity writes only the chain-sourced fields; ErrAgentNotInitialized if agent.Id has no row.
	UpdateOnChainIdentity(ctx context.Context, agent *models.Agent) error
}
