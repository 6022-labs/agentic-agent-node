package utils

import (
	"github.com/6022-labs/agentic-agent-node/src/agent_reasoning/models"
	"github.com/samber/lo"
)

// ToolNames lists the tool names a provider is allowed to call when it talks to an MCP server directly.
func ToolNames(tools []*models.Tool) []string {
	return lo.Map(tools, func(tool *models.Tool, _ int) string { return tool.Name })
}
