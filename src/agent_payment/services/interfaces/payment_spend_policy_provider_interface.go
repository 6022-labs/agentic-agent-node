package interfaces

import (
	"context"

	"github.com/6022-labs/agentic-agent-node/src/agent_payment/models"
)

// PaymentSpendPolicyProviderInterface materializes the agent's spend policy head on first access, with no rule.
// A service, not a use case, so the signer's evaluator can depend on it.
type PaymentSpendPolicyProviderInterface interface {
	// Ensure returns the policy head, creating it if absent.
	Ensure(ctx context.Context) (*models.PaymentSpendPolicy, error)
}
