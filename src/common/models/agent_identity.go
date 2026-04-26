package models

// AgentIdentity is the on-chain identity of an agent, resolved via the agent
// collections manager + the collection's ERC721 metadata. Resolved lazily by
// AgentIdentityResolver; absent until an NFT has been minted to the agent's
// wallet.
type AgentIdentity struct {
	CollectionAddress string
	TokenId           uint64
	Name              string
}
