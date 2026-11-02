package repositories

import (
	"context"

	"github.com/6022-labs/agentic-agent-node/src/agent_payment/models"
	"github.com/google/uuid"
)

// PaymentSpendRuleWindowRepositoryInterface persists a spend rule's recurring budgets and their running spend.
type PaymentSpendRuleWindowRepositoryInterface interface {
	// ListByRuleId returns the rule's windows, shortest period first.
	ListByRuleId(ctx context.Context, ruleId uuid.UUID) ([]models.PaymentSpendRuleWindow, error)
	// ListByRuleIdForUpdate reads the rule's windows locked for a caller that will record spend on them.
	ListByRuleIdForUpdate(ctx context.Context, ruleId uuid.UUID) ([]models.PaymentSpendRuleWindow, error)
	Create(ctx context.Context, window *models.PaymentSpendRuleWindow) (*models.PaymentSpendRuleWindow, error)
	Update(ctx context.Context, window *models.PaymentSpendRuleWindow) (*models.PaymentSpendRuleWindow, error)
	DeleteById(ctx context.Context, id uuid.UUID) error
}
