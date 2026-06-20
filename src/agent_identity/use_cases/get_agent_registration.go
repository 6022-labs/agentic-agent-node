package use_cases

import (
	"context"

	"github.com/6022-labs/agentic-agent-node/src/agent_identity/repositories"
	"github.com/6022-labs/agentic-agent-node/src/agent_identity/use_cases/responses"
	"github.com/6022-labs/agentic-agent-node/src/common/errors"
)

// GetAgentRegistration returns the stored draft + status, used to pre-fill the
// form on retry after a failed mint or ENS registration.
type GetAgentRegistration struct {
	agentRepository                  repositories.AgentRepositoryInterface
	agentRegistrationStateRepository repositories.AgentRegistrationStateRepositoryInterface
}

func NewGetAgentRegistration(
	agentRepository repositories.AgentRepositoryInterface,
	agentRegistrationStateRepository repositories.AgentRegistrationStateRepositoryInterface,
) *GetAgentRegistration {
	return &GetAgentRegistration{agentRepository: agentRepository, agentRegistrationStateRepository: agentRegistrationStateRepository}
}

func (u *GetAgentRegistration) Execute(ctx context.Context) (*responses.AgentRegistrationResponse, error) {
	agent, err := u.agentRepository.Get(ctx)
	if err != nil {
		return nil, errors.NewInternalError("agent_lookup_failed", err.Error())
	}
	if agent == nil {
		return nil, errors.NewNotFoundError("draft_not_found", "no agent registration has been saved yet")
	}

	state, err := u.agentRegistrationStateRepository.GetByAgentId(ctx, agent.Id)
	if err != nil {
		return nil, errors.NewInternalError("state_lookup_failed", err.Error())
	}

	resp := &responses.AgentRegistrationResponse{}
	resp.FromModel(agent, state)
	return resp, nil
}
