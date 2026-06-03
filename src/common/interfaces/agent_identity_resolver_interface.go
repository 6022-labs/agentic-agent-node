package interfaces

import (
	"context"
	"errors"

	"github.com/6022-labs/agentic-agent-node/src/common/models"
)

// ErrIdentityNotFound means no NFT is minted to the agent's wallet yet. Both
// the resolver and the gateway HTTP client return it so callers map it
// consistently (e.g. to a 503).
var ErrIdentityNotFound = errors.New("agent identity not found on chain")

type AgentIdentityResolverInterface interface {
	Resolve(ctx context.Context) (*models.AgentIdentity, error)
}
