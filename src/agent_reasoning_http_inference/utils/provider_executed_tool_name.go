package utils

import (
	"github.com/6022-labs/agentic-agent-node/src/agent_reasoning/models"
	reasoning_utils "github.com/6022-labs/agentic-agent-node/src/agent_reasoning/utils"
)

// ProviderExecutedToolFunctionName rebuilds mcp_<conn>_<tool> for a known echoed server label, else keeps the raw name.
func ProviderExecutedToolFunctionName(
	mcpTools map[*models.MCPConnection][]*models.Tool,
	serverLabel, toolName string,
) string {
	mcpConnection, found := reasoning_utils.FindMcpConnectionByName(mcpTools, serverLabel)
	if !found {
		return toolName
	}
	return reasoning_utils.McpToolToFunctionName(mcpConnection, &models.Tool{Name: toolName})
}
