package models

// ExternalChainConfig is one chain's config from the runtime.
type ExternalChainConfig struct {
	CollectionsManagerAddress string  `json:"collectionsManagerAddress"`
	CollectionCreatorAddress  string  `json:"collectionCreatorAddress"`
	Token6022Address          string  `json:"token6022Address"`
	EnsRegistryAddress        string  `json:"ensRegistryAddress"`
	Mintable                  bool    `json:"mintable"`
	FacilitatorNetwork        *string `json:"facilitatorNetwork"`
}

// ExternalBlockchainConfigPayload mirrors the runtime blockchain config response.
type ExternalBlockchainConfigPayload struct {
	Chains map[string]ExternalChainConfig `json:"chains"`
}
