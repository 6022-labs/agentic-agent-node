package interfaces

import "github.com/ethereum/go-ethereum/common"

// AgentCollectionsManagerRequesterInterface enumerates the agent collections
// registered on chain. Used during bootstrap to discover which collection
// minted this agent's NFT.
type AgentCollectionsManagerRequesterInterface interface {
	// ListAllCollections returns every registered collection address, in the
	// order the manager stores them.
	ListAllCollections() ([]common.Address, error)
}
