package utils

import (
	"github.com/6022-labs/agentic-agent-node/src/agent_reasoning/models"
	reasoning_utils "github.com/6022-labs/agentic-agent-node/src/agent_reasoning/utils"
	"github.com/samber/lo"
)

// ProviderExecutedToolFunctionName rebuilds the local mcp_<conn>_<tool> name for
// a tool a provider ran itself, but only when the echoed server label matches a
// known connection; otherwise it keeps the provider's raw name (no fabrication).
func ProviderExecutedToolFunctionName(mcpTools map[*models.MCPConnection][]*models.Tool, serverLabel, toolName string) string {
	mcpConnection, found := lo.Find(lo.Keys(mcpTools), func(mcpConnection *models.MCPConnection) bool {
		return mcpConnection.Name == serverLabel
	})
	if !found {
		return toolName
	}
	return reasoning_utils.McpToolToFunctionName(mcpConnection, &models.Tool{Name: toolName})
}
