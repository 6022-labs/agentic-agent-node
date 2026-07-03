package responses

// EvmChainConfig is one EVM chain's contract addresses, whether it's mintable (a
// creator is deployed), and its x402 facilitator network — null when the chain
// has an RPC but no facilitator.
type EvmChainConfig struct {
	CollectionsManagerAddress string  `json:"collectionsManagerAddress"`
	CollectionCreatorAddress  string  `json:"collectionCreatorAddress"`
	Token6022Address          string  `json:"token6022Address"`
	EnsRegistryAddress        string  `json:"ensRegistryAddress"`
	Mintable                  bool    `json:"mintable"`
	FacilitatorNetwork        *string `json:"facilitatorNetwork"`
}

// BlockchainConfigResponse is the reachable EVM chains keyed by chain id, which
// the owner UI boots from.
type BlockchainConfigResponse struct {
	EvmChains map[string]EvmChainConfig `json:"evmChains"`
}
