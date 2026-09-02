package interfaces

import (
	"context"

	identity_models "github.com/6022-labs/agentic-agent-node/src/agent_identity/models"
	"github.com/6022-labs/agentic-agent-node/src/agent_payment/models"
	"github.com/google/uuid"
)

// PaymentReceivePolicyProviderInterface materializes the receive policy on first access,
// seeding the owner's free rule. Requires a mint: the seed prices access to a token.
type PaymentReceivePolicyProviderInterface interface {
	// Ensure returns the policy head and the agent it belongs to, creating the policy if absent.
	Ensure(ctx context.Context) (*models.PaymentReceivePolicy, *identity_models.Agent, error)
	// EvmAddress returns the agent's stored evm address, or "" when none exists.
	EvmAddress(ctx context.Context, agentId uuid.UUID) (string, error)
}
