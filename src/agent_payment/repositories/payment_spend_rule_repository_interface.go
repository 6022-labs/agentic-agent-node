package repositories

import (
	"context"

	"github.com/6022-labs/agentic-agent-node/src/agent_payment/models"
	"github.com/google/uuid"
)

// PaymentSpendRuleRepositoryInterface persists the spend policy's rules.
// Position is display order and tie-break only, never authorization.
type PaymentSpendRuleRepositoryInterface interface {
	// ListByPolicyId returns the policy's rules, ordered by position.
	ListByPolicyId(ctx context.Context, policyId uuid.UUID) ([]models.PaymentSpendRule, error)
	// GetById returns the rule, or (nil, nil) if it does not exist.
	GetById(ctx context.Context, id uuid.UUID) (*models.PaymentSpendRule, error)
	// Create appends the rule at the end of the table order.
	Create(ctx context.Context, rule *models.PaymentSpendRule) (*models.PaymentSpendRule, error)
	Update(ctx context.Context, rule *models.PaymentSpendRule) (*models.PaymentSpendRule, error)
	DeleteById(ctx context.Context, id uuid.UUID) error
}
