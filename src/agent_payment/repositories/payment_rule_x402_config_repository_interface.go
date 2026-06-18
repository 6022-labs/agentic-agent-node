package repositories

import (
	"context"

	"github.com/6022-labs/agentic-agent-node/src/agent_payment/models"
	"github.com/google/uuid"
)

// PaymentRuleX402ConfigRepositoryInterface persists the x402 config row of a
// rule, keyed by the rule id (1:1). The x402 strategy owns it; the generic
// payment domain never sees it.
type PaymentRuleX402ConfigRepositoryInterface interface {
	// Upsert replaces the rule's x402 config.
	Upsert(ctx context.Context, config *models.PaymentRuleX402Config) error
	// GetByRuleId returns the config, or (nil, nil) if none exists.
	GetByRuleId(ctx context.Context, ruleId uuid.UUID) (*models.PaymentRuleX402Config, error)
}
