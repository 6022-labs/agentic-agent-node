package interfaces

import "context"

// AgentCardSignature is one A2A detached JWS: base64url protected header and signature.
type AgentCardSignature struct {
	Protected string
	Signature string
}

// AgentCardSignerInterface produces an A2A §8.4 detached JWS over an agent card.
type AgentCardSignerInterface interface {
	// Sign JCS-canonicalizes card and signs it with the node's EVM wallet key.
	// card must not carry a "signatures" member — the spec excludes it from the signed content.
	Sign(ctx context.Context, card []byte) (*AgentCardSignature, error)
}
