package repositories

import (
	"context"

	"github.com/6022-labs/agentic-agent-node/src/agent_payment/models"
	"github.com/google/uuid"
)

// PaymentReceivePolicyRepositoryInterface persists the policy head. One policy per
// agent, so an agent holds zero or one row.
type PaymentReceivePolicyRepositoryInterface interface {
	// GetByAgentId returns the agent's policy head, or (nil, nil) if none exists.
	GetByAgentId(ctx context.Context, agentId uuid.UUID) (*models.PaymentReceivePolicy, error)
	Create(ctx context.Context, policy *models.PaymentReceivePolicy) (*models.PaymentReceivePolicy, error)
	// DeleteByAgentId removes the agent's policy head and all its rules. No-op if none is set.
	DeleteByAgentId(ctx context.Context, agentId uuid.UUID) error
}
