package repositories

import (
	"context"

	"github.com/6022-labs/agentic-agent-node/src/agent_payment/models"
	"github.com/google/uuid"
)

// PaymentReceiveRulePriceRepositoryInterface persists a rule's prices, at most one per amount.
type PaymentReceiveRulePriceRepositoryInterface interface {
	// ListByRuleId returns the rule's prices, oldest first.
	ListByRuleId(ctx context.Context, ruleId uuid.UUID) ([]models.PaymentReceiveRulePrice, error)
	Create(ctx context.Context, price *models.PaymentReceiveRulePrice) (*models.PaymentReceiveRulePrice, error)
	// DeleteByRuleId removes the rule's prices; no-op if none.
	DeleteByRuleId(ctx context.Context, ruleId uuid.UUID) error
}
