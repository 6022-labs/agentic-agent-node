package repositories

import (
	"context"

	"github.com/6022-labs/agentic-agent-node/src/agent_payment/models"
	"github.com/google/uuid"
)

// PaymentSpendPolicyRepositoryInterface persists the spend policy head. One
// policy per agent, so an agent holds zero or one row.
type PaymentSpendPolicyRepositoryInterface interface {
	// GetByAgentId returns the agent's spend policy head, or (nil, nil) if none exists.
	GetByAgentId(ctx context.Context, agentId uuid.UUID) (*models.PaymentSpendPolicy, error)
	Create(ctx context.Context, policy *models.PaymentSpendPolicy) (*models.PaymentSpendPolicy, error)
	Update(ctx context.Context, policy *models.PaymentSpendPolicy) (*models.PaymentSpendPolicy, error)
	// DeleteByAgentId removes the agent's spend policy head and all its rules. No-op if none is set.
	DeleteByAgentId(ctx context.Context, agentId uuid.UUID) error
}
