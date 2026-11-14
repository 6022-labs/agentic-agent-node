package use_cases

import (
	"context"

	"github.com/6022-labs/agentic-agent-node/src/agent_identity/repositories"
	identity_interfaces "github.com/6022-labs/agentic-agent-node/src/agent_identity/services/interfaces"
	"github.com/6022-labs/agentic-agent-node/src/agent_identity/use_cases/responses"
	"github.com/6022-labs/agentic-agent-node/src/common/errors"
	"github.com/samber/lo"
)

type GetAgentIdentity struct {
	agentResolver                    identity_interfaces.AgentResolverInterface
	agentRegistrationStateRepository repositories.AgentRegistrationStateRepositoryInterface
}

func NewGetAgentIdentity(
	agentResolver identity_interfaces.AgentResolverInterface,
	agentRegistrationStateRepository repositories.AgentRegistrationStateRepositoryInterface,
) *GetAgentIdentity {
	return &GetAgentIdentity{
		agentResolver:                    agentResolver,
		agentRegistrationStateRepository: agentRegistrationStateRepository,
	}
}

func (u *GetAgentIdentity) Execute(ctx context.Context) (*responses.GetAgentIdentityResponse, error) {
	agent, err := u.agentResolver.Get(ctx)
	if err != nil {
		return nil, errors.NewInternalError("identity_lookup_failed", err)
	}
	if agent == nil || agent.CollectionTokenId == nil {
		return nil, errors.NewNotFoundError("identity_not_found", "agent identity not found")
	}

	state, err := u.agentRegistrationStateRepository.GetByAgentId(ctx, agent.Id)
	if err != nil {
		return nil, errors.NewInternalError("identity_lookup_failed", err)
	}
	if !state.IsMinted() {
		return nil, errors.NewNotFoundError("identity_not_found", "agent identity not found")
	}

	return &responses.GetAgentIdentityResponse{
		CollectionAddress: agent.CollectionAddress,
		TokenId:           *agent.CollectionTokenId,
		Name:              lo.FromPtr(agent.Name),
	}, nil
}
