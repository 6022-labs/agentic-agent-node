package responses

import "github.com/6022-labs/agentic-agent-node/src/agent_identity/models"

// AgentRegistrationResponse is the stored draft + current on-chain status,
// used both as the save result and to pre-fill the form on retry.
type AgentRegistrationResponse struct {
	ChainId           string            `json:"chainId"`
	CollectionAddress string            `json:"collectionAddress"`
	TokenId           *uint64           `json:"tokenId"`
	ProposalId        *uint64           `json:"proposalId"`
	Name              string            `json:"name"`
	Role              string            `json:"role"`
	CloneOf           *uint64           `json:"cloneOf"`
	Owner             string            `json:"owner"`
	WalletAddress     string            `json:"walletAddress"`
	EnsDomain         string            `json:"ensDomain"`
	Images            map[string]string `json:"images"`
	Attributes        map[string]string `json:"attributes"`
	MintStatus        string            `json:"mintStatus"`
	EnsStatus         string            `json:"ensStatus"`
}

func (r *AgentRegistrationResponse) FromModel(agent *models.Agent, state *models.AgentRegistrationState) {
	owner := ""
	if agent.Owner != nil {
		owner = *agent.Owner
	}
	r.ChainId = agent.ChainId
	r.CollectionAddress = agent.CollectionAddress
	r.TokenId = agent.CollectionTokenId
	r.ProposalId = agent.ProposalId
	r.Name = agent.Name
	r.Role = agent.Role.String()
	r.CloneOf = agent.CloneOf
	r.Owner = owner
	r.WalletAddress = agent.WalletAddress
	r.EnsDomain = agent.EnsDomain
	r.Images = agent.Images
	r.Attributes = agent.Attributes
	if state != nil {
		r.MintStatus = state.MintStatus.String()
		r.EnsStatus = state.EnsRegistryStatus.String()
	}
}
