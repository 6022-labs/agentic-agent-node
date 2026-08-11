package repositories

import (
	"context"

	"github.com/6022-labs/agentic-agent-node/src/agent_reasoning/models"
	"github.com/google/uuid"
)

// SelfToolRepositoryInterface persists the agent's per-tool enabled flags.
// Rows are reconciled against the registry by SelfToolReconciler.
type SelfToolRepositoryInterface interface {
	ListByAgentId(ctx context.Context, agentId uuid.UUID) ([]models.SelfTool, error)
	GetById(ctx context.Context, agentId uuid.UUID, id uuid.UUID) (*models.SelfTool, error)
	Create(ctx context.Context, tool *models.SelfTool) (*models.SelfTool, error)
	Update(ctx context.Context, tool *models.SelfTool) (*models.SelfTool, error)

	// ListExistingNames returns which of the given names already have a row for the agent.
	ListExistingNames(ctx context.Context, agentId uuid.UUID, names []string) ([]string, error)

	// DeleteAllExceptNames removes the agent's rows whose name is not in the given set, returning the count removed.
	DeleteAllExceptNames(ctx context.Context, agentId uuid.UUID, names []string) (int64, error)
}
