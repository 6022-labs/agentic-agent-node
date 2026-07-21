package responses

import "github.com/6022-labs/agentic-agent-node/src/agent_identity/models/enums"

// KeyValue mirrors the on-chain AgentCollectionV1/AgentEnsRegistry KeyValue tuple in contract argument order.
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

// MintInfoResponse is the exact calldata for AgentCollectionV1.mint/createMintProposal plus moderator state.
type MintInfoResponse struct {
	CollectionAddress string                       `json:"collectionAddress"`
	To                string                       `json:"to"`
	Name              string                       `json:"name"`
	Addresses         map[enums.AddressType]string `json:"addresses"`
	Images            []KeyValue                   `json:"images"`
	Attributes        []KeyValue                   `json:"attributes"`
	CloneOf           *uint64                      `json:"cloneOf"`
	ModeratorCount    uint64                       `json:"moderatorCount"`
	OwnerIsModerator  bool                         `json:"ownerIsModerator"`
}
