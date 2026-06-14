package models

// ExternalAgentCollectionsPayload mirrors the runtime collections listing response.
type ExternalAgentCollectionsPayload struct {
	Collections []ExternalAgentCollectionPayload `json:"collections"`
}

// ExternalAgentCollectionPayload mirrors one collection from the on-chain manager.
type ExternalAgentCollectionPayload struct {
	Address        string `json:"address"`
	Name           string `json:"name"`
	Description    string `json:"description"`
	ModeratorCount uint64 `json:"moderatorCount"`
	NextTokenId    uint64 `json:"nextTokenId"`
}
