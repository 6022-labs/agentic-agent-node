package responses

// BlockchainConfigResponse is the chain id and contract addresses the owner UI
// needs at boot. Static for the life of the process.
type BlockchainConfigResponse struct {
	ChainId                   uint64 `json:"chainId"`
	CollectionsManagerAddress string `json:"collectionsManagerAddress"`
	CollectionCreatorAddress  string `json:"collectionCreatorAddress"`
	Token6022Address          string `json:"token6022Address"`
	EnsRegistryAddress        string `json:"ensRegistryAddress"`
}
