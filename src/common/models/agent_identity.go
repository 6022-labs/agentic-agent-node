package models

// AgentIdentity is an agent's on-chain identity, resolved lazily by
// AgentIdentityResolver from the collections manager + ERC721 metadata. Absent
// until an NFT is minted to the agent's wallet.
type AgentIdentity struct {
	CollectionAddress string
	TokenId           uint64
	Name              string
}
