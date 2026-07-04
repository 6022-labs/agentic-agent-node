package use_cases

import (
	"context"

	identity_repositories "github.com/6022-labs/agentic-agent-node/src/agent_identity/repositories"
	"github.com/6022-labs/agentic-agent-node/src/agent_reasoning/repositories"
	"github.com/6022-labs/agentic-agent-node/src/agent_reasoning/use_cases/responses"
	"github.com/6022-labs/agentic-agent-node/src/common/errors"
)

type GetAgentGuidance struct {
	*AbstractReasoningUseCase
	agentGuidanceRepository repositories.AgentGuidanceRepositoryInterface
}

func NewGetAgentGuidance(
	agentGuidanceRepository repositories.AgentGuidanceRepositoryInterface,
	agentRepository identity_repositories.AgentRepositoryInterface,
) *GetAgentGuidance {
	return &GetAgentGuidance{
		AbstractReasoningUseCase: NewAbstractReasoningUseCase(agentRepository),
		agentGuidanceRepository:  agentGuidanceRepository,
	}
}

// Execute returns the guidance content; empty when the owner never set one.
func (u *GetAgentGuidance) Execute(ctx context.Context) (*responses.AgentGuidanceResponse, error) {
	agent, err := u.resolveAgent(ctx)
	if err != nil {
		return nil, err
	}
	guidance, err := u.agentGuidanceRepository.GetByAgentId(ctx, agent.Id)
	if err != nil {
		return nil, errors.NewInternalError("guidance_lookup_failed", err.Error())
	}
	response := &responses.AgentGuidanceResponse{}
	if guidance != nil {
		response.Content = guidance.Content
	}
	return response, nil
}
