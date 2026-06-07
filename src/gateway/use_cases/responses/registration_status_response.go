package responses

// RegistrationStatusResponse is the gateway-shaped mint + ENS status, returned
// by the check and self-register endpoints.
type RegistrationStatusResponse struct {
	MintStatus string  `json:"mintStatus"`
	EnsStatus  string  `json:"ensStatus"`
	TokenId    *uint64 `json:"tokenId"`
	ProposalId *uint64 `json:"proposalId"`
	MintTxHash *string `json:"mintTxHash"`
	EnsTxHash  *string `json:"ensTxHash"`
}
