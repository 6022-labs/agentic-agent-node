package use_cases

import (
	"context"

	"github.com/6022-labs/agentic-agent-node/src/agent_identity/repositories"
	"github.com/6022-labs/agentic-agent-node/src/agent_identity/use_cases/responses"
	"github.com/6022-labs/agentic-agent-node/src/common/errors"
)

type GetAgentIdentity struct {
	agentRepository repositories.AgentRepositoryInterface
	stateRepository repositories.AgentRegistrationStateRepositoryInterface
}

func NewGetAgentIdentity(
	agentRepository repositories.AgentRepositoryInterface,
	stateRepository repositories.AgentRegistrationStateRepositoryInterface,
) *GetAgentIdentity {
	return &GetAgentIdentity{agentRepository: agentRepository, stateRepository: stateRepository}
}

func (u *GetAgentIdentity) Execute(ctx context.Context) (*responses.GetAgentIdentityResponse, error) {
	state, err := u.stateRepository.Get(ctx)
	if err != nil {
		return nil, errors.NewInternalError("identity_lookup_failed", err.Error())
	}
	if !state.IsMinted() {
		return nil, errors.NewNotFoundError("identity_not_found", "agent identity not found")
	}

	agent, err := u.agentRepository.Get(ctx)
	if err != nil {
		return nil, errors.NewInternalError("identity_lookup_failed", err.Error())
	}
	if agent == nil || agent.CollectionTokenId == nil {
		return nil, errors.NewNotFoundError("identity_not_found", "agent identity not found")
	}

	return &responses.GetAgentIdentityResponse{
		CollectionAddress: agent.CollectionAddress,
		TokenId:           *agent.CollectionTokenId,
		Name:              agent.Name,
	}, nil
}
