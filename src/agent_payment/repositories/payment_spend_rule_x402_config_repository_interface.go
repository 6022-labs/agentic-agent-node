package repositories

import (
	"context"

	"github.com/6022-labs/agentic-agent-node/src/agent_payment/models"
	"github.com/google/uuid"
)

// PaymentSpendRuleX402ConfigRepositoryInterface persists the x402 token of a
// spend rule, keyed by the rule id (1:1). The x402 strategy owns it.
type PaymentSpendRuleX402ConfigRepositoryInterface interface {
	// Upsert replaces the rule's x402 token.
	Upsert(ctx context.Context, config *models.PaymentSpendRuleX402Config) error
	// GetByRuleId returns the rule's token, or (nil, nil) if none exists.
	GetByRuleId(ctx context.Context, ruleId uuid.UUID) (*models.PaymentSpendRuleX402Config, error)
	// GetByRuleIdForUpdate is GetByRuleId holding the row until the scope ends.
	GetByRuleIdForUpdate(ctx context.Context, ruleId uuid.UUID) (*models.PaymentSpendRuleX402Config, error)
	// GetByPolicyIdAndAsset returns the policy's config for that exact network and lowercased asset, or (nil, nil).
	GetByPolicyIdAndAsset(
		ctx context.Context,
		policyId uuid.UUID,
		network string,
		asset string,
	) (*models.PaymentSpendRuleX402Config, error)
}
