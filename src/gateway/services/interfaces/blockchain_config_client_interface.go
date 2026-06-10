package interfaces

import (
	"context"

	"github.com/6022-labs/agentic-agent-node/src/gateway/models"
)

// BlockchainConfigClientInterface is the gateway-side port for the runtime's
// static chain configuration. Implemented over HTTP.
type BlockchainConfigClientInterface interface {
	// Get returns the runtime's chain id and contract addresses.
	Get(ctx context.Context) (*models.ExternalBlockchainConfigPayload, error)
}
