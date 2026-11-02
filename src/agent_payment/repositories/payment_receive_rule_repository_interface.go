package repositories

import (
	"context"

	"github.com/6022-labs/agentic-agent-node/src/agent_payment/models"
	"github.com/google/uuid"
)

// PaymentReceiveRuleRepositoryInterface persists the policy's pricing rules, one per token.
type PaymentReceiveRuleRepositoryInterface interface {
	// ListByPolicyId returns the policy's rules, oldest first.
	ListByPolicyId(ctx context.Context, policyId uuid.UUID) ([]models.PaymentReceiveRule, error)
	// GetById returns the rule, or (nil, nil) if it does not exist.
	GetById(ctx context.Context, id uuid.UUID) (*models.PaymentReceiveRule, error)
	Create(ctx context.Context, rule *models.PaymentReceiveRule) (*models.PaymentReceiveRule, error)
	Update(ctx context.Context, rule *models.PaymentReceiveRule) (*models.PaymentReceiveRule, error)
	DeleteById(ctx context.Context, id uuid.UUID) error
}
