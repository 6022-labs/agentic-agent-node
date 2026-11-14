package use_cases

import (
	"context"

	identity_interfaces "github.com/6022-labs/agentic-agent-node/src/agent_identity/services/interfaces"
	"github.com/6022-labs/agentic-agent-node/src/agent_reasoning/repositories"
	"github.com/6022-labs/agentic-agent-node/src/agent_reasoning/services/interfaces"
	"github.com/6022-labs/agentic-agent-node/src/agent_reasoning/use_cases/requests"
	"github.com/6022-labs/agentic-agent-node/src/agent_reasoning/use_cases/responses"
	"github.com/6022-labs/agentic-agent-node/src/common/errors"
)

type UpdateSelfTool struct {
	*AbstractReasoningUseCase
	selfToolRepository repositories.SelfToolRepositoryInterface
	selfToolRegistry   interfaces.SelfToolRegistryInterface
}

func NewUpdateSelfTool(
	selfToolRepository repositories.SelfToolRepositoryInterface,
	selfToolRegistry interfaces.SelfToolRegistryInterface,
	agentResolver identity_interfaces.AgentResolverInterface,
) *UpdateSelfTool {
	return &UpdateSelfTool{
		AbstractReasoningUseCase: NewAbstractReasoningUseCase(agentResolver),
		selfToolRepository:       selfToolRepository,
		selfToolRegistry:         selfToolRegistry,
	}
}

func (u *UpdateSelfTool) Execute(
	ctx context.Context,
	request *requests.UpdateSelfToolRequest,
) (*responses.SelfToolResponse, error) {
	if err := request.ValidateAndSanitize(); err != nil {
		return nil, err
	}

	agent, err := u.resolveAgent(ctx)
	if err != nil {
		return nil, err
	}

	tool, err := u.selfToolRepository.GetById(ctx, agent.Id, request.Id)
	if err != nil {
		return nil, errors.NewInternalError("get_failed", err)
	}
	if tool == nil {
		return nil, errors.NewNotFoundError("self_tool_not_found", "self tool not found")
	}

	tool.Enabled = *request.Enabled
	updated, err := u.selfToolRepository.Update(ctx, tool)
	if err != nil {
		return nil, errors.NewInternalError("update_failed", err)
	}

	description := ""
	for _, definition := range u.selfToolRegistry.ListTools() {
		if definition.Name == updated.Name {
			description = definition.Description
			break
		}
	}
	return responses.ToSelfToolResponse(updated, description, u.selfToolRegistry.ListGroupNames()[updated.Name]), nil
}
