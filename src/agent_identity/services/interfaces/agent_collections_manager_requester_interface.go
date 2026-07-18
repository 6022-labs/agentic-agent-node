package interfaces

import (
	"github.com/6022-labs/agentic-agent-node/src/agent_identity/models"
	"github.com/6022-labs/agentic-agent-node/src/agent_identity/models/enums"
	"github.com/ethereum/go-ethereum/common"
)

// AgentCollectionsManagerRequesterInterface reads the on-chain collections
// manager to find which collection minted this agent's NFT.
type AgentCollectionsManagerRequesterInterface interface {
	// FindAgentByAddress returns the agent NFT a typed address belongs to, or a zero collection if none.
	FindAgentByAddress(chainId string, addressType enums.AddressType, value string) (collection common.Address, tokenId uint64, name string, err error)
	// IsKnownCollection reports whether the manager governs this collection.
	IsKnownCollection(chainId string, collection common.Address) (bool, error)
	// ListCollections returns every collection the manager governs, exhausting its paginated listing.
	ListCollections(chainId string) ([]models.AgentCollection, error)
}
