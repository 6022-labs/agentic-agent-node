package repositories

import (
	"context"

	"github.com/6022-labs/agentic-agent-node/src/agent_payment/models"
)

// PaymentPolicyRepositoryInterface persists the policy head. One agent per
// container, so the table holds zero or one row today.
type PaymentPolicyRepositoryInterface interface {
	// Get returns the policy head, or (nil, nil) if none exists.
	Get(ctx context.Context) (*models.PaymentPolicy, error)
	Create(ctx context.Context, policy *models.PaymentPolicy) (*models.PaymentPolicy, error)
	// Delete removes the policy head and all its rules. No-op if none is set.
	Delete(ctx context.Context) error
}
