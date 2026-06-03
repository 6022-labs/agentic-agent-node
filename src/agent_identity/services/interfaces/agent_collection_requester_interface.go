package interfaces

import "github.com/ethereum/go-ethereum/common"

// AgentCollectionRequesterInterface reads from an ERC721 agent collection. The
// collection address is passed per call so one requester serves any collection.
type AgentCollectionRequesterInterface interface {
	// TokenIdByWallet returns the tokenId minted to this wallet, or 0 if none.
	TokenIdByWallet(collectionAddress common.Address, wallet common.Address) (uint64, error)
	// OwnerOf returns the current NFT owner for this token.
	OwnerOf(collectionAddress common.Address, tokenId uint64) (common.Address, error)
	// NameOf returns the on-chain name (ENS prefix) for this token.
	NameOf(collectionAddress common.Address, tokenId uint64) (string, error)
}
