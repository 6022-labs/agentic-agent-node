package repositories

import (
	"context"

	"github.com/6022-labs/agentic-agent-node/src/agent_payment/models"
	runtime_common_repositories "github.com/6022-labs/agentic-agent-node/src/runtime_common/repositories"
	"github.com/google/uuid"
)

// PaymentSpendRuleWindowRepositoryInterface persists a spend rule's recurring budgets and their running spend.
type PaymentSpendRuleWindowRepositoryInterface interface {
	// ListByRuleId returns the rule's windows, shortest period first.
	ListByRuleId(
		ctx context.Context,
		ruleId uuid.UUID,
		options ...runtime_common_repositories.ReadOption,
	) ([]models.PaymentSpendRuleWindow, error)
	Create(ctx context.Context, window *models.PaymentSpendRuleWindow) (*models.PaymentSpendRuleWindow, error)
	Update(ctx context.Context, window *models.PaymentSpendRuleWindow) (*models.PaymentSpendRuleWindow, error)
	DeleteById(ctx context.Context, id uuid.UUID) error
}
