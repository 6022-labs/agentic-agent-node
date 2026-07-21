package interfaces

import (
	"github.com/6022-labs/agentic-agent-node/src/agent_identity/models/enums"
	"github.com/ethereum/go-ethereum/common"
)

// OnChainAgentInfo is the subset of AgentCollectionV1.informationOf the node
// reconciles into its local agent row after a mint completes.
type OnChainAgentInfo struct {
	Name       string
	Creator    common.Address
	CloneOf    *uint64
	Images     map[string]string
	Attributes map[string]string
}

// AgentCollectionRequesterInterface reads from an ERC721 agent collection. The
// chain id and collection address are passed per call so one requester serves
// any collection on any reachable chain.
type AgentCollectionRequesterInterface interface {
	// OwnerOf returns the current NFT owner for this token.
	OwnerOf(chainId string, collectionAddress common.Address, tokenId uint64) (common.Address, error)
	// CollectionName returns the ERC721 collection-level name.
	CollectionName(chainId string, collectionAddress common.Address) (string, error)
	// AddressToTokenId returns the tokenId a typed address maps to, or 0 if none.
	AddressToTokenId(chainId string, collectionAddress common.Address, addressType enums.AddressType, value string) (uint64, error)
	// NameToTokenId returns the tokenId an agent name maps to, or 0 if unused.
	NameToTokenId(chainId string, collectionAddress common.Address, name string) (uint64, error)
	// InformationOf returns the on-chain agent record for a minted token.
	InformationOf(chainId string, collectionAddress common.Address, tokenId uint64) (OnChainAgentInfo, error)
	// ModeratorCount returns how many moderators gate direct minting (0 = open).
	ModeratorCount(chainId string, collectionAddress common.Address) (uint64, error)
	// IsModerator reports whether account may mint directly when moderators exist.
	IsModerator(chainId string, collectionAddress common.Address, account common.Address) (bool, error)
}
