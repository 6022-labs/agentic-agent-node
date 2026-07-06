package interfaces

import (
	"context"
	identity_responses "github.com/6022-labs/agentic-agent-node/src/agent_identity/use_cases/responses"
)

// CheckAgentMintExecutorInterface exposes the idempotent mint reconciliation to the self tools.
type CheckAgentMintExecutorInterface interface {
	Execute(ctx context.Context) (*identity_responses.RegistrationStatusResponse, error)
}
