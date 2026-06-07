package interfaces

import "github.com/ethereum/go-ethereum/common"

// OnChainAgentInfo is the subset of AgentCollectionV1.informationOf the node
// reconciles into its local agent row after a mint completes.
type OnChainAgentInfo struct {
	Name    string
	Role    string
	Creator common.Address
	CloneOf *uint64
}

// AgentCollectionRequesterInterface reads from an ERC721 agent collection. The
// collection address is passed per call so one requester serves any collection.
type AgentCollectionRequesterInterface interface {
	// OwnerOf returns the current NFT owner for this token.
	OwnerOf(collectionAddress common.Address, tokenId uint64) (common.Address, error)
	// CollectionName returns the ERC721 collection-level name.
	CollectionName(collectionAddress common.Address) (string, error)
	// WalletToTokenId returns the tokenId a controlled wallet maps to, or 0 if none.
	WalletToTokenId(collectionAddress common.Address, wallet common.Address) (uint64, error)
	// NameToTokenId returns the tokenId an agent name maps to, or 0 if unused.
	NameToTokenId(collectionAddress common.Address, name string) (uint64, error)
	// InformationOf returns the on-chain agent record for a minted token.
	InformationOf(collectionAddress common.Address, tokenId uint64) (OnChainAgentInfo, error)
	// ModeratorCount returns how many moderators gate direct minting (0 = open).
	ModeratorCount(collectionAddress common.Address) (uint64, error)
	// IsModerator reports whether account may mint directly when moderators exist.
	IsModerator(collectionAddress common.Address, account common.Address) (bool, error)
}
