package responses

// AgentCardSignature is one A2A detached JWS over the canonicalized card, minus this member.
type AgentCardSignature struct {
	Protected string `json:"protected"`
	Signature string `json:"signature"`
}
