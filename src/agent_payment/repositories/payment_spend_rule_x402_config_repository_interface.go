package repositories

import (
	"context"

	"github.com/6022-labs/agentic-agent-node/src/agent_payment/models"
	"github.com/google/uuid"
)

// PaymentSpendRuleX402ConfigRepositoryInterface persists the x402 matcher of a
// spend rule, keyed by the rule id (1:1). The x402 strategy owns it.
type PaymentSpendRuleX402ConfigRepositoryInterface interface {
	// Upsert replaces the rule's x402 matcher.
	Upsert(ctx context.Context, config *models.PaymentSpendRuleX402Config) error
	// GetByRuleId returns the matcher, or (nil, nil) if none exists.
	GetByRuleId(ctx context.Context, ruleId uuid.UUID) (*models.PaymentSpendRuleX402Config, error)
}
