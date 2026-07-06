package interfaces

import (
	"context"

	"github.com/6022-labs/agentic-agent-node/src/agent_reasoning/models"
)

type McpRequesterInterface interface {
	ListTools(ctx context.Context, mcpConnection *models.MCPConnection) ([]*models.Tool, error)
	CallTool(
		ctx context.Context,
		mcpConnection *models.MCPConnection,
		tool *models.Tool,
		toolInput map[string]any,
	) (map[string]any, error)
}
