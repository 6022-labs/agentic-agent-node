package responses

import "github.com/6022-labs/agentic-agent-node/src/agent_identity/models"

// RegistrationStatusResponse is returned by the check + self-register endpoints.
type RegistrationStatusResponse struct {
	MintStatus string  `json:"mintStatus"`
	EnsStatus  string  `json:"ensStatus"`
	TokenId    *uint64 `json:"tokenId"`
	ProposalId *uint64 `json:"proposalId"`
	MintTxHash *string `json:"mintTxHash"`
	EnsTxHash  *string `json:"ensTxHash"`
}

func (r *RegistrationStatusResponse) FromModel(agent *models.Agent, state *models.AgentRegistrationState) {
	r.MintStatus = state.MintStatus.String()
	r.EnsStatus = state.EnsRegistryStatus.String()
	if agent != nil {
		r.TokenId = agent.CollectionTokenId
		r.ProposalId = agent.ProposalId
	}
	if state != nil {
		r.MintTxHash = state.MintTxHash
		r.EnsTxHash = state.EnsTxHash
	}
}
