package responses

// ChainConfig is one chain's config surfaced to the owner UI.
type ChainConfig struct {
	CollectionsManagerAddress string  `json:"collectionsManagerAddress"`
	CollectionCreatorAddress  string  `json:"collectionCreatorAddress"`
	Token6022Address          string  `json:"token6022Address"`
	EnsRegistryAddress        string  `json:"ensRegistryAddress"`
	Mintable                  bool    `json:"mintable"`
	FacilitatorNetwork        *string `json:"facilitatorNetwork"`
}

// BlockchainConfigResponse is the public chain config (chains keyed by chain id)
// the owner UI boots from.
type BlockchainConfigResponse struct {
	Chains map[string]ChainConfig `json:"chains"`
}
