package interfaces

import (
	"context"
	identity_responses "github.com/6022-labs/agentic-agent-node/src/agent_identity/use_cases/responses"
)

// GetAgentRegistrationExecutorInterface exposes the registration read to the self tools.
type GetAgentRegistrationExecutorInterface interface {
	Execute(ctx context.Context) (*identity_responses.AgentRegistrationResponse, error)
}
