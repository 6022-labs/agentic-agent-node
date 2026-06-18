package responses

type GetAgentIdentityResponse struct {
	CollectionAddress string `json:"collectionAddress"`
	TokenId           uint64 `json:"tokenId"`
	Name              string `json:"name"`
}
