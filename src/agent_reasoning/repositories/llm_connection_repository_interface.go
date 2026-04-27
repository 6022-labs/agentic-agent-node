package repositories

import "github.com/6022-labs/agentic-agent-node/src/agent_reasoning/models"

// LLMConnectionRepositoryInterface persists the agent's LLM provider config.
// One agent per container, so the table holds zero or one row at a time.
type LLMConnectionRepositoryInterface interface {
	// Get returns the configured connection, or (nil, nil) if none has been set.
	Get() (*models.LLMConnection, error)
	// Upsert creates the row if absent, otherwise updates the existing row in
	// place (preserving Id + CreatedAt). Returns the persisted entity.
	Upsert(connection *models.LLMConnection) (*models.LLMConnection, error)
	// Delete clears the configured connection. No-op if none is set.
	Delete() error
}
