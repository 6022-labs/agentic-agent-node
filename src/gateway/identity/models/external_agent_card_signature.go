package models

// ExternalAgentCardSignaturePayload is the runtime's detached JWS over an agent card.
type ExternalAgentCardSignaturePayload struct {
	Protected string `json:"protected"`
	Signature string `json:"signature"`
}
