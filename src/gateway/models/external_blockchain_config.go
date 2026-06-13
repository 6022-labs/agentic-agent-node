package models

// ExternalBlockchainConfigPayload mirrors the runtime blockchain config response.
type ExternalBlockchainConfigPayload struct {
	ChainId                   uint64   `json:"chainId"`
	CollectionsManagerAddress string   `json:"collectionsManagerAddress"`
	CollectionCreatorAddress  string   `json:"collectionCreatorAddress"`
	Token6022Address          string   `json:"token6022Address"`
	EnsRegistryAddress        string   `json:"ensRegistryAddress"`
	FacilitatorNetworks       []string `json:"facilitatorNetworks"`
}
