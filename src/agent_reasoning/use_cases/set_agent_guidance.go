package use_cases

import (
	"context"

	identity_repositories "github.com/6022-labs/agentic-agent-node/src/agent_identity/repositories"
	"github.com/6022-labs/agentic-agent-node/src/agent_reasoning/models"
	"github.com/6022-labs/agentic-agent-node/src/agent_reasoning/repositories"
	"github.com/6022-labs/agentic-agent-node/src/agent_reasoning/use_cases/requests"
	"github.com/6022-labs/agentic-agent-node/src/agent_reasoning/use_cases/responses"
	"github.com/6022-labs/agentic-agent-node/src/common/errors"
)

type SetAgentGuidance struct {
	*AbstractReasoningUseCase
	agentGuidanceRepository repositories.AgentGuidanceRepositoryInterface
}

func NewSetAgentGuidance(
	agentGuidanceRepository repositories.AgentGuidanceRepositoryInterface,
	agentRepository identity_repositories.AgentRepositoryInterface,
) *SetAgentGuidance {
	return &SetAgentGuidance{
		AbstractReasoningUseCase: NewAbstractReasoningUseCase(agentRepository),
		agentGuidanceRepository:  agentGuidanceRepository,
	}
}

func (u *SetAgentGuidance) Execute(
	ctx context.Context,
	request *requests.SetAgentGuidanceRequest,
) (*responses.AgentGuidanceResponse, error) {
	agent, err := u.resolveAgent(ctx)
	if err != nil {
		return nil, err
	}
	saved, err := u.agentGuidanceRepository.Upsert(ctx, &models.AgentGuidance{
		AgentId: agent.Id,
		Content: request.Content,
	})
	if err != nil {
		return nil, errors.NewInternalError("guidance_update_failed", err.Error())
	}
	return &responses.AgentGuidanceResponse{Content: saved.Content}, nil
}
