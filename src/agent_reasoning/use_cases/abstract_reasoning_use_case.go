package use_cases

import (
	"context"

	identity_models "github.com/6022-labs/agentic-agent-node/src/agent_identity/models"
	identity_repositories "github.com/6022-labs/agentic-agent-node/src/agent_identity/repositories"
	"github.com/6022-labs/agentic-agent-node/src/common/errors"
)

// AbstractReasoningUseCase carries the agent lookup every reasoning use case
// needs to scope its connection rows.
type AbstractReasoningUseCase struct {
	agentRepository identity_repositories.AgentRepositoryInterface
}

func NewAbstractReasoningUseCase(
	agentRepository identity_repositories.AgentRepositoryInterface,
) *AbstractReasoningUseCase {
	return &AbstractReasoningUseCase{
		agentRepository: agentRepository,
	}
}

// resolveAgent loads the single agent identity, failing when it has not been
// bootstrapped yet. The returned agent's Id scopes the runtime DB rows.
func (u *AbstractReasoningUseCase) resolveAgent(ctx context.Context) (*identity_models.Agent, error) {
	agent, err := u.agentRepository.Get(ctx)
	if err != nil {
		return nil, errors.NewInternalError("agent_lookup_failed", err)
	}
	if agent == nil {
		return nil, errors.NewInternalError("agent_lookup_failed", errors.ErrAgentNotInitialized)
	}
	return agent, nil
}
