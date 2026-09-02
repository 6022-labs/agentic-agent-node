package interfaces

import (
	"context"

	"github.com/6022-labs/agentic-agent-node/src/agent_payment/models"
)

// PaymentSpendPolicyProviderInterface materializes the agent's spend policy on first access, seeding a
// catch-all rule. A service, not a use case, so the signer's evaluator can depend on it.
type PaymentSpendPolicyProviderInterface interface {
	// Ensure returns the policy head, creating it with the seeded rule if absent.
	Ensure(ctx context.Context) (*models.PaymentSpendPolicy, error)
}
