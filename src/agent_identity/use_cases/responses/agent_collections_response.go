package responses

// AgentCollectionsResponse lists the collections the on-chain manager governs.
type AgentCollectionsResponse struct {
	Collections []AgentCollectionResponse `json:"collections"`
}

// AgentCollectionResponse describes one mintable collection for the owner UI.
type AgentCollectionResponse struct {
	Address        string `json:"address"`
	Name           string `json:"name"`
	Description    string `json:"description"`
	ModeratorCount uint64 `json:"moderatorCount"`
	NextTokenId    uint64 `json:"nextTokenId"`
}
