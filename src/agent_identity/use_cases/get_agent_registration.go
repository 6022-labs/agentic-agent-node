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
	agentAddressRepository           repositories.AgentAddressRepositoryInterface
	agentRegistrationStateRepository repositories.AgentRegistrationStateRepositoryInterface
}

func NewGetAgentRegistration(
	agentRepository repositories.AgentRepositoryInterface,
	agentAddressRepository repositories.AgentAddressRepositoryInterface,
	agentRegistrationStateRepository repositories.AgentRegistrationStateRepositoryInterface,
) *GetAgentRegistration {
	return &GetAgentRegistration{
		agentRepository:                  agentRepository,
		agentAddressRepository:           agentAddressRepository,
		agentRegistrationStateRepository: agentRegistrationStateRepository,
	}
}

func (u *GetAgentRegistration) Execute(ctx context.Context) (*responses.AgentRegistrationResponse, error) {
	agent, err := u.agentRepository.Get(ctx)
	if err != nil {
		return nil, errors.NewInternalError("agent_lookup_failed", err)
	}
	if agent == nil {
		return nil, errors.NewNotFoundError("draft_not_found", "no agent registration has been saved yet")
	}

	state, err := u.agentRegistrationStateRepository.GetByAgentId(ctx, agent.Id)
	if err != nil {
		return nil, errors.NewInternalError("state_lookup_failed", err)
	}
	// No state row means only the empty bootstrap agent exists; a real draft or mint always seeds one.
	if state == nil {
		return nil, errors.NewNotFoundError("draft_not_found", "no agent registration has been saved yet")
	}

	addresses, err := u.agentAddressRepository.GetAllByAgentId(ctx, agent.Id)
	if err != nil {
		return nil, errors.NewInternalError("agent_addresses_lookup_failed", err)
	}

	resp := &responses.AgentRegistrationResponse{}
	resp.FromModel(agent, addresses, state)
	return resp, nil
}
