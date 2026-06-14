package responses

// EnsInfoResponse is the gateway-shaped initAgentTexts calldata.
type EnsInfoResponse struct {
	CollectionAddress string     `json:"collectionAddress"`
	TokenId           uint64     `json:"tokenId"`
	Records           []KeyValue `json:"records"`
}
