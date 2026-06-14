package responses

// AgentCollectionsResponse lists the collections the on-chain manager governs.
type AgentCollectionsResponse struct {
	Collections []AgentCollectionResponse `json:"collections"`
}

// AgentCollectionResponse describes one collection the owner can mint into.
type AgentCollectionResponse struct {
	Address        string `json:"address"`
	Name           string `json:"name"`
	Description    string `json:"description"`
	ModeratorCount uint64 `json:"moderatorCount"`
	NextTokenId    uint64 `json:"nextTokenId"`
}
