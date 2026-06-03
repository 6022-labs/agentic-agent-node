package interfaces

import "github.com/ethereum/go-ethereum/common"

// AgentCollectionRequesterInterface reads from an ERC721 agent collection. The
// collection address is passed per call so one requester serves any collection.
type AgentCollectionRequesterInterface interface {
	// OwnerOf returns the current NFT owner for this token.
	OwnerOf(collectionAddress common.Address, tokenId uint64) (common.Address, error)
	// CollectionName returns the ERC721 collection-level name.
	CollectionName(collectionAddress common.Address) (string, error)
}
