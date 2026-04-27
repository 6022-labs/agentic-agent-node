package repositories

import (
	"github.com/6022-labs/agentic-agent-node/src/agent_reasoning/models"
	"github.com/google/uuid"
)

// MCPConnectionRepositoryInterface persists the agent's configured MCP servers.
// An agent may have many.
type MCPConnectionRepositoryInterface interface {
	List() ([]models.MCPConnection, error)
	GetById(id uuid.UUID) (*models.MCPConnection, error)
	Create(connection *models.MCPConnection) (*models.MCPConnection, error)
	Update(connection *models.MCPConnection) (*models.MCPConnection, error)
	Delete(id uuid.UUID) error
}
