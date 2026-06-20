package repositories

import (
	"context"

	"github.com/6022-labs/agentic-agent-node/src/agent_reasoning/models"
	"github.com/google/uuid"
)

// MCPConnectionRepositoryInterface persists the agent's configured MCP servers.
// An agent may have many.
type MCPConnectionRepositoryInterface interface {
	ListByAgentId(ctx context.Context, agentId uuid.UUID) ([]models.MCPConnection, error)
	GetById(ctx context.Context, agentId uuid.UUID, id uuid.UUID) (*models.MCPConnection, error)
	Create(ctx context.Context, connection *models.MCPConnection) (*models.MCPConnection, error)
	Update(ctx context.Context, connection *models.MCPConnection) (*models.MCPConnection, error)
	DeleteById(ctx context.Context, agentId uuid.UUID, id uuid.UUID) error
}
