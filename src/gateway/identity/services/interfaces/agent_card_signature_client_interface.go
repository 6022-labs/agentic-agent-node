package interfaces

import (
	"context"

	"github.com/6022-labs/agentic-agent-node/src/gateway/identity/models"
)

// AgentCardSignatureClientInterface asks the runtime — the only host holding the wallet key — to sign an agent card.
type AgentCardSignatureClientInterface interface {
	// card must not carry a "signatures" member.
	Sign(ctx context.Context, card []byte) (*models.ExternalAgentCardSignaturePayload, error)
}
