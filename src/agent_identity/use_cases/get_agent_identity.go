package use_cases

import (
	"context"

	"github.com/6022-labs/agentic-agent-node/src/agent_identity/repositories"
	"github.com/6022-labs/agentic-agent-node/src/agent_identity/use_cases/responses"
	"github.com/6022-labs/agentic-agent-node/src/common/usecase_result"
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

func (u *GetAgentIdentity) Execute(ctx context.Context) *usecase_result.Result {
	state, err := u.stateRepository.Get(ctx)
	if err != nil {
		return usecase_result.InternalError("identity_lookup_failed", err.Error())
	}
	if !state.IsMinted() {
		return usecase_result.NotFound("identity_not_found", "agent identity not found")
	}

	agent, err := u.agentRepository.Get(ctx)
	if err != nil {
		return usecase_result.InternalError("identity_lookup_failed", err.Error())
	}
	if agent == nil || agent.CollectionTokenId == nil {
		return usecase_result.NotFound("identity_not_found", "agent identity not found")
	}

	return usecase_result.Ok(&responses.GetAgentIdentityResponse{
		CollectionAddress: agent.CollectionAddress,
		TokenId:           *agent.CollectionTokenId,
		Name:              agent.Name,
	})
}
