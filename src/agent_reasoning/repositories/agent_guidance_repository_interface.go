package repositories

import (
	"context"

	"github.com/6022-labs/agentic-agent-node/src/agent_reasoning/models"
	"github.com/google/uuid"
)

// AgentGuidanceRepositoryInterface persists the single owner-editable guidance
// row per agent.
type AgentGuidanceRepositoryInterface interface {
	// GetByAgentId returns the agent's guidance row, or (nil, nil) when unset.
	GetByAgentId(ctx context.Context, agentId uuid.UUID) (*models.AgentGuidance, error)

	// Upsert keeps a single row per agent: update in place if one exists, otherwise insert.
	Upsert(ctx context.Context, guidance *models.AgentGuidance) (*models.AgentGuidance, error)
}
