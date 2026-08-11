package tools

import (
	"context"

	"github.com/6022-labs/agentic-agent-node/src/agent_reasoning/models"
	reasoning_requests "github.com/6022-labs/agentic-agent-node/src/agent_reasoning/use_cases/requests"
	tool_interfaces "github.com/6022-labs/agentic-agent-node/src/agent_reasoning_self_tools/tools/interfaces"
	"github.com/6022-labs/agentic-agent-node/src/agent_reasoning_self_tools/utils"
)

// DeleteMCPConnectionTool exposes the DeleteMCPConnection use case as a built-in self tool.
type DeleteMCPConnectionTool struct {
	deleteMCPConnectionExecutor tool_interfaces.DeleteMCPConnectionExecutorInterface
}

func NewDeleteMCPConnectionTool(
	deleteMCPConnectionExecutor tool_interfaces.DeleteMCPConnectionExecutorInterface,
) *DeleteMCPConnectionTool {
	return &DeleteMCPConnectionTool{deleteMCPConnectionExecutor: deleteMCPConnectionExecutor}
}

func (t *DeleteMCPConnectionTool) Definition() *models.Tool {
	return &models.Tool{
		Name:        "delete_mcp_connection",
		Description: "Disconnect yourself from an MCP server. Its tools disappear on your next turn.",
		InputScheme: utils.ObjectScheme(map[string]models.ToolInputSchemeProperty{
			"id": utils.StringProperty("Id of the connection to delete (see list_mcp_connections)."),
		}, []string{"id"}),
	}
}

func (t *DeleteMCPConnectionTool) Execute(ctx context.Context, args map[string]any) (map[string]any, error) {
	id, err := utils.UuidArg(args, "id")
	if err != nil {
		return nil, err
	}
	if err := t.deleteMCPConnectionExecutor.Execute(ctx, &reasoning_requests.DeleteMCPConnectionRequest{Id: id}); err != nil {
		return nil, err
	}
	return utils.SuccessResult(), nil
}

func (t *DeleteMCPConnectionTool) IsEnabledByDefault() bool {
	return true
}
