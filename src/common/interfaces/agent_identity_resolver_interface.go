package interfaces

import (
	"context"
	"errors"

	"github.com/6022-labs/agentic-agent-node/src/common/models"
)

// ErrIdentityNotFound means no NFT has been minted to the agent's wallet yet.
// Both the in-process resolver and the gateway-side HTTP client return this
// sentinel so callers can map it consistently (e.g. to a 503 / "not yet").
var ErrIdentityNotFound = errors.New("agent identity not found on chain")

type AgentIdentityResolverInterface interface {
	Resolve(ctx context.Context) (*models.AgentIdentity, error)
}
