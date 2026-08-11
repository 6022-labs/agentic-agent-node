package tools

import (
	"context"

	"github.com/6022-labs/agentic-agent-node/src/agent_reasoning/models"
	tool_interfaces "github.com/6022-labs/agentic-agent-node/src/agent_reasoning_self_tools/tools/interfaces"
	"github.com/6022-labs/agentic-agent-node/src/agent_reasoning_self_tools/utils"
)

// ListMCPConnectionsTool exposes the ListMCPConnections use case as a built-in self tool.
type ListMCPConnectionsTool struct {
	listMCPConnectionsExecutor tool_interfaces.ListMCPConnectionsExecutorInterface
}

func NewListMCPConnectionsTool(listMCPConnectionsExecutor tool_interfaces.ListMCPConnectionsExecutorInterface) *ListMCPConnectionsTool {
	return &ListMCPConnectionsTool{listMCPConnectionsExecutor: listMCPConnectionsExecutor}
}

func (t *ListMCPConnectionsTool) Definition() *models.Tool {
	return &models.Tool{
		Name:        "list_mcp_connections",
		Description: "List your MCP server connections (name, url, auth header names only).",
		InputScheme: utils.EmptyObjectScheme(),
	}
}

func (t *ListMCPConnectionsTool) Execute(ctx context.Context, _ map[string]any) (map[string]any, error) {
	return utils.ToToolResult(t.listMCPConnectionsExecutor.Execute(ctx))
}

func (t *ListMCPConnectionsTool) IsEnabledByDefault() bool {
	return true
}
