package responses

// AgentRegistrationResponse is the gateway-shaped draft + on-chain status,
// returned by the save and get-draft endpoints.
type AgentRegistrationResponse struct {
	ChainId           string            `json:"chainId"`
	CollectionAddress string            `json:"collectionAddress"`
	TokenId           *uint64           `json:"tokenId"`
	ProposalId        *uint64           `json:"proposalId"`
	Name              string            `json:"name"`
	Role              string            `json:"role"`
	CloneOf           *uint64           `json:"cloneOf"`
	Owner             string            `json:"owner"`
	Addresses         map[string]string `json:"addresses"`
	EnsDomain         string            `json:"ensDomain"`
	Images            map[string]string `json:"images"`
	Attributes        map[string]string `json:"attributes"`
	MintStatus        string            `json:"mintStatus"`
	EnsStatus         string            `json:"ensStatus"`
}
