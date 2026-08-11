package tools

import (
	"context"

	"github.com/6022-labs/agentic-agent-node/src/agent_reasoning/models"
	reasoning_requests "github.com/6022-labs/agentic-agent-node/src/agent_reasoning/use_cases/requests"
	tool_interfaces "github.com/6022-labs/agentic-agent-node/src/agent_reasoning_self_tools/tools/interfaces"
	"github.com/6022-labs/agentic-agent-node/src/agent_reasoning_self_tools/utils"
)

// CreateMCPConnectionTool exposes CreateMCPConnection; no header-auth input, so no secret transits the model.
type CreateMCPConnectionTool struct {
	createMCPConnectionExecutor tool_interfaces.CreateMCPConnectionExecutorInterface
}

func NewCreateMCPConnectionTool(
	createMCPConnectionExecutor tool_interfaces.CreateMCPConnectionExecutorInterface,
) *CreateMCPConnectionTool {
	return &CreateMCPConnectionTool{createMCPConnectionExecutor: createMCPConnectionExecutor}
}

func (t *CreateMCPConnectionTool) Definition() *models.Tool {
	return &models.Tool{
		Name:        "create_mcp_connection",
		Description: "Connect yourself to a new MCP server (auth-less only). Its tools become available on your next turn.",
		InputScheme: utils.ObjectScheme(map[string]models.ToolInputSchemeProperty{
			"name": utils.StringProperty("Connection name (no underscores, \"self\" is reserved)."),
			"url":  utils.StringProperty("The MCP server url."),
		}, []string{"name", "url"}),
	}
}

func (t *CreateMCPConnectionTool) Execute(ctx context.Context, args map[string]any) (map[string]any, error) {
	name, err := utils.StringArg(args, "name")
	if err != nil {
		return nil, err
	}
	url, err := utils.StringArg(args, "url")
	if err != nil {
		return nil, err
	}
	return utils.ToToolResult(t.createMCPConnectionExecutor.Execute(ctx, &reasoning_requests.CreateMCPConnectionRequest{
		Name: name,
		Url:  url,
		// The agent picks the server, not a tool subset — the owner narrows it from the console.
		AllowedToolPatterns: []string{".*"},
	}))
}

func (t *CreateMCPConnectionTool) IsEnabledByDefault() bool {
	return true
}

func (t *CreateMCPConnectionTool) GroupName() string {
	return GroupMCP
}
