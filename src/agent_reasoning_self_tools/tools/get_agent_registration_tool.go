package tools

import (
	"context"
	"github.com/6022-labs/agentic-agent-node/src/agent_reasoning/models"
	tool_interfaces "github.com/6022-labs/agentic-agent-node/src/agent_reasoning_self_tools/tools/interfaces"
	"github.com/6022-labs/agentic-agent-node/src/agent_reasoning_self_tools/utils"
)

// GetAgentRegistrationTool exposes the GetAgentRegistration use case as a built-in self tool.
type GetAgentRegistrationTool struct {
	getAgentRegistrationExecutor tool_interfaces.GetAgentRegistrationExecutorInterface
}

func NewGetAgentRegistrationTool(getAgentRegistrationExecutor tool_interfaces.GetAgentRegistrationExecutorInterface) *GetAgentRegistrationTool {
	return &GetAgentRegistrationTool{getAgentRegistrationExecutor: getAgentRegistrationExecutor}
}

func (t *GetAgentRegistrationTool) Definition() *models.Tool {
	return &models.Tool{
		Name:        "get_registration",
		Description: "Read your registration draft and its mint/ENS status.",
		InputScheme: utils.EmptyObjectScheme(),
	}
}

func (t *GetAgentRegistrationTool) Execute(ctx context.Context, _ map[string]any) (map[string]any, error) {
	return utils.ToToolResult(t.getAgentRegistrationExecutor.Execute(ctx))
}
