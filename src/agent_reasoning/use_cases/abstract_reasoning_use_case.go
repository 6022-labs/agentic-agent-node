package use_cases

import (
	"context"

	identity_models "github.com/6022-labs/agentic-agent-node/src/agent_identity/models"
	identity_interfaces "github.com/6022-labs/agentic-agent-node/src/agent_identity/services/interfaces"
	"github.com/6022-labs/agentic-agent-node/src/common/errors"
)

// AbstractReasoningUseCase carries the agent lookup every reasoning use case
// needs to scope its connection rows.
type AbstractReasoningUseCase struct {
	agentResolver identity_interfaces.AgentResolverInterface
}

func NewAbstractReasoningUseCase(
	agentResolver identity_interfaces.AgentResolverInterface,
) *AbstractReasoningUseCase {
	return &AbstractReasoningUseCase{
		agentResolver: agentResolver,
	}
}

// resolveAgent loads the single agent identity, failing when it has not been
// bootstrapped yet. The returned agent's Id scopes the runtime DB rows.
func (u *AbstractReasoningUseCase) resolveAgent(ctx context.Context) (*identity_models.Agent, error) {
	agent, err := u.agentResolver.Get(ctx)
	if err != nil {
		return nil, errors.NewInternalError("agent_lookup_failed", err)
	}
	if agent == nil {
		return nil, errors.NewInternalError("agent_lookup_failed", errors.ErrAgentNotInitialized)
	}
	return agent, nil
}
