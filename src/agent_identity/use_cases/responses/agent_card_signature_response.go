package responses

import identity_interfaces "github.com/6022-labs/agentic-agent-node/src/agent_identity/services/interfaces"

type AgentCardSignatureResponse struct {
	Protected string `json:"protected"`
	Signature string `json:"signature"`
}

func (r *AgentCardSignatureResponse) FromModel(signature *identity_interfaces.AgentCardSignature) {
	r.Protected = signature.Protected
	r.Signature = signature.Signature
}
