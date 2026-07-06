package tools

import (
	"context"
	"github.com/6022-labs/agentic-agent-node/src/agent_reasoning/models"
	reasoning_requests "github.com/6022-labs/agentic-agent-node/src/agent_reasoning/use_cases/requests"
	tool_interfaces "github.com/6022-labs/agentic-agent-node/src/agent_reasoning_self_tools/tools/interfaces"
	"github.com/6022-labs/agentic-agent-node/src/agent_reasoning_self_tools/utils"
)

// SetAgentGuidanceTool exposes the SetAgentGuidance use case as a built-in self tool.
type SetAgentGuidanceTool struct {
	setAgentGuidanceExecutor tool_interfaces.SetAgentGuidanceExecutorInterface
}

func NewSetAgentGuidanceTool(setAgentGuidanceExecutor tool_interfaces.SetAgentGuidanceExecutorInterface) *SetAgentGuidanceTool {
	return &SetAgentGuidanceTool{setAgentGuidanceExecutor: setAgentGuidanceExecutor}
}

func (t *SetAgentGuidanceTool) Definition() *models.Tool {
	return &models.Tool{
		Name:        "set_guidance",
		Description: "Replace the owner guidance layered into your system prompt. Takes effect on your next turn.",
		InputScheme: utils.ObjectScheme(map[string]models.ToolInputSchemeProperty{
			"content": utils.StringProperty("The full new guidance text."),
		}, []string{"content"}),
	}
}

func (t *SetAgentGuidanceTool) Execute(ctx context.Context, args map[string]any) (map[string]any, error) {
	content, err := utils.StringArg(args, "content")
	if err != nil {
		return nil, err
	}
	return utils.ToToolResult(t.setAgentGuidanceExecutor.Execute(ctx, &reasoning_requests.SetAgentGuidanceRequest{Content: content}))
}
