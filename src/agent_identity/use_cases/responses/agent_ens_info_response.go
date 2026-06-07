package responses

// EnsInfoResponse is the calldata for AgentEnsRegistry.initAgentTexts.
type EnsInfoResponse struct {
	CollectionAddress string     `json:"collectionAddress"`
	TokenId           uint64     `json:"tokenId"`
	Records           []KeyValue `json:"records"`
}
