package interfaces

import (
	"github.com/6022-labs/agentic-agent-node/src/agent_reasoning/models"
)

type McpRequesterInterface interface {
	ListTools(mcpConnection *models.MCPConnection) ([]*models.Tool, error)
	CallTool(
		mcpConnection *models.MCPConnection,
		tool *models.Tool,
		toolInput map[string]any,
	) (map[string]any, error)
}
