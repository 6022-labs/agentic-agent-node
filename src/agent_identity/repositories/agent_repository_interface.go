package repositories

import "github.com/6022-labs/agentic-agent-node/src/agent_identity/models"

// AgentRepositoryInterface persists the agent's identity as a single row.
type AgentRepositoryInterface interface {
	// Get returns the persisted agent, or (nil, nil) if none has been set.
	Get() (*models.Agent, error)
	Upsert(agent *models.Agent) (*models.Agent, error)
}
