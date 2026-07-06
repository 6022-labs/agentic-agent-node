package interfaces

import (
	"context"
	identity_responses "github.com/6022-labs/agentic-agent-node/src/agent_identity/use_cases/responses"
)

// CheckAgentEnsExecutorInterface exposes the idempotent ENS reconciliation to the self tools.
type CheckAgentEnsExecutorInterface interface {
	Execute(ctx context.Context) (*identity_responses.RegistrationStatusResponse, error)
}
