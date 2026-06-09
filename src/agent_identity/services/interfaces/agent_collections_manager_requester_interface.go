package interfaces

import (
	"github.com/6022-labs/agentic-agent-node/src/agent_identity/models"
	"github.com/ethereum/go-ethereum/common"
)

// AgentCollectionsManagerRequesterInterface reads the on-chain collections
// manager to find which collection minted this agent's NFT.
type AgentCollectionsManagerRequesterInterface interface {
	// FindAgentByWallet returns the agent NFT the wallet controls, or a zero collection if none.
	FindAgentByWallet(wallet common.Address) (collection common.Address, tokenId uint64, name string, err error)
	// IsKnownCollection reports whether the manager governs this collection.
	IsKnownCollection(collection common.Address) (bool, error)
	// ListCollections returns every collection the manager governs, exhausting its paginated listing.
	ListCollections() ([]models.AgentCollection, error)
}
