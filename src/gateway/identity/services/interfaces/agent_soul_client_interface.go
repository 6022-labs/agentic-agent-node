package interfaces

import (
	"context"
	"errors"

	"github.com/6022-labs/agentic-agent-node/src/gateway/identity/models"
)

// ErrAgentSoulNotFound — runtime reported no agent identity to carry a soul.
var ErrAgentSoulNotFound = errors.New("agent soul not found")

// AgentSoulClientInterface is the gateway-side port for the runtime's agent
// soul (immutable mint-seeded persona) endpoint. Implemented over HTTP.
type AgentSoulClientInterface interface {
	Get(ctx context.Context) (*models.ExternalAgentSoulPayload, error)
}
