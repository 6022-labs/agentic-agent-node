package utils

import (
	"github.com/6022-labs/agentic-agent-node/src/agent_reasoning/models"
	reasoning_utils "github.com/6022-labs/agentic-agent-node/src/agent_reasoning/utils"
)

// ProviderExecutedToolFunctionName rebuilds the local mcp_<conn>_<tool> name when the echoed server label matches a known connection, else keeps the provider's raw name.
func ProviderExecutedToolFunctionName(mcpTools map[*models.MCPConnection][]*models.Tool, serverLabel, toolName string) string {
	mcpConnection, found := reasoning_utils.FindMcpConnectionByName(mcpTools, serverLabel)
	if !found {
		return toolName
	}
	return reasoning_utils.McpToolToFunctionName(mcpConnection, &models.Tool{Name: toolName})
}
