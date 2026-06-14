package interfaces

import (
	"context"

	"github.com/6022-labs/agentic-agent-node/src/gateway/identity/models"
)

// AgentCollectionsClientInterface is the gateway-side port for the runtime's
// on-chain collections catalog. Implemented over HTTP.
type AgentCollectionsClientInterface interface {
	// List returns every collection the on-chain manager governs.
	List(ctx context.Context) (*models.ExternalAgentCollectionsPayload, error)
}
