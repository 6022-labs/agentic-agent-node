package responses

// EvmChainConfig is one EVM chain's config surfaced to the owner UI.
type EvmChainConfig struct {
	CollectionsManagerAddress string  `json:"collectionsManagerAddress"`
	CollectionCreatorAddress  string  `json:"collectionCreatorAddress"`
	Token6022Address          string  `json:"token6022Address"`
	EnsRegistryAddress        string  `json:"ensRegistryAddress"`
	Mintable                  bool    `json:"mintable"`
	FacilitatorNetwork        *string `json:"facilitatorNetwork"`
}

// BlockchainConfigResponse is the public EVM chain config (keyed by chain id)
// the owner UI boots from.
type BlockchainConfigResponse struct {
	EvmChains map[string]EvmChainConfig `json:"evmChains"`
}
