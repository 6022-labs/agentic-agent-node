package interfaces

import (
	"context"

	"github.com/6022-labs/agentic-agent-node/src/gateway/reasoning/models"
)

// AgentGuidanceClientInterface is the gateway-side port for the runtime's
// owner-editable guidance endpoint. Implemented over HTTP.
type AgentGuidanceClientInterface interface {
	Get(ctx context.Context) (*models.ExternalAgentGuidancePayload, error)
	Set(ctx context.Context, body *models.ExternalAgentGuidancePayload) (*models.ExternalAgentGuidancePayload, error)
}
