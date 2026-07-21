package responses

// KeyValue mirrors a contract KeyValue tuple in the order a client must pass it.
type KeyValue struct {
	Key   string `json:"key"`
	Value string `json:"value"`
}

// MintInfoResponse is the gateway-shaped mint() calldata + moderator state.
type MintInfoResponse struct {
	CollectionAddress string            `json:"collectionAddress"`
	To                string            `json:"to"`
	Name              string            `json:"name"`
	Addresses         map[string]string `json:"addresses"`
	Images            []KeyValue        `json:"images"`
	Attributes        []KeyValue        `json:"attributes"`
	CloneOf           *uint64           `json:"cloneOf"`
	ModeratorCount    uint64            `json:"moderatorCount"`
	OwnerIsModerator  bool              `json:"ownerIsModerator"`
}
