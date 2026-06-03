package interfaces

import "github.com/ethereum/go-ethereum/common"

// AgentCollectionsManagerRequesterInterface enumerates on-chain agent
// collections, used at bootstrap to find which minted this agent's NFT.
type AgentCollectionsManagerRequesterInterface interface {
	// ListAllCollections returns every registered collection address, in store order.
	ListAllCollections() ([]common.Address, error)
}
