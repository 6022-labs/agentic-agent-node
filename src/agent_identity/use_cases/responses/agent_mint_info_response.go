package responses

// KeyValue mirrors the on-chain AgentCollectionV1/AgentEnsRegistry KeyValue
// tuple in the order a client must pass it to the contract.
type KeyValue struct {
	Key   string `json:"key"`
	Value string `json:"value"`
}

// FromMap builds a KeyValue slice from a string map.
func (KeyValue) FromMap(m map[string]string) []KeyValue {
	out := make([]KeyValue, 0, len(m))
	for k, v := range m {
		out = append(out, KeyValue{Key: k, Value: v})
	}
	return out
}

// MintInfoResponse is the exact calldata for AgentCollectionV1.mint /
// createMintProposal, plus the moderator state the client uses to pick which.
type MintInfoResponse struct {
	CollectionAddress string     `json:"collectionAddress"`
	To                string     `json:"to"`
	Role              string     `json:"role"`
	Name              string     `json:"name"`
	Wallets           []string   `json:"wallets"`
	Images            []KeyValue `json:"images"`
	Attributes        []KeyValue `json:"attributes"`
	CloneOf           *uint64    `json:"cloneOf"`
	ModeratorCount    uint64     `json:"moderatorCount"`
	OwnerIsModerator  bool       `json:"ownerIsModerator"`
}
