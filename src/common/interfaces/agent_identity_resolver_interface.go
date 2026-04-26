package interfaces

import (
	"context"

	"github.com/6022-labs/agentic-agent-node/src/common/models"
)

// AgentIdentityResolverInterface lazily resolves the agent's on-chain identity
// (collection + tokenId + name). Caches on success.
//
// Implementations should return a sentinel "not found" error (see the
// agent_runtime services package) when no NFT has been minted to the agent's
// wallet yet, so callers can map that case to a 503.
type AgentIdentityResolverInterface interface {
	Resolve(ctx context.Context) (*models.AgentIdentity, error)
}
