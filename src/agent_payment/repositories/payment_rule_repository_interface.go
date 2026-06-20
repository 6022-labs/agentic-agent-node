package repositories

import (
	"context"

	"github.com/6022-labs/agentic-agent-node/src/agent_payment/models"
	"github.com/google/uuid"
)

// PaymentRuleRepositoryInterface persists the policy's pricing rules.
type PaymentRuleRepositoryInterface interface {
	// ListByPolicyId returns the policy's rules, ordered by position.
	ListByPolicyId(ctx context.Context, policyId uuid.UUID) ([]models.PaymentRule, error)
	// GetById returns the rule, or (nil, nil) if it does not exist.
	GetById(ctx context.Context, id uuid.UUID) (*models.PaymentRule, error)
	// Create appends the rule at the end of the table order.
	Create(ctx context.Context, rule *models.PaymentRule) (*models.PaymentRule, error)
	Update(ctx context.Context, rule *models.PaymentRule) (*models.PaymentRule, error)
	DeleteById(ctx context.Context, id uuid.UUID) error
}
