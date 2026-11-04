package repositories

import (
	"context"

	"github.com/6022-labs/agentic-agent-node/src/agent_payment/models"
	runtime_common_repositories "github.com/6022-labs/agentic-agent-node/src/runtime_common/repositories"
	"github.com/google/uuid"
)

// PaymentSpendRuleRepositoryInterface persists the spend policy's rules, one per token.
type PaymentSpendRuleRepositoryInterface interface {
	// ListByPolicyId returns the policy's rules, oldest first.
	ListByPolicyId(ctx context.Context, policyId uuid.UUID) ([]models.PaymentSpendRule, error)
	// GetById returns the rule, or (nil, nil) if it does not exist.
	GetById(
		ctx context.Context,
		id uuid.UUID,
		options ...runtime_common_repositories.ReadOption,
	) (*models.PaymentSpendRule, error)
	Create(ctx context.Context, rule *models.PaymentSpendRule) (*models.PaymentSpendRule, error)
	Update(ctx context.Context, rule *models.PaymentSpendRule) (*models.PaymentSpendRule, error)
	DeleteById(ctx context.Context, id uuid.UUID) error
}
