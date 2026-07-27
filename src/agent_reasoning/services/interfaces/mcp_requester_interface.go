package interfaces

import (
	"context"
	"errors"

	"github.com/6022-labs/agentic-agent-node/src/agent_reasoning/models"
)

// ErrMCPUnauthorized — the MCP server refused the connection's auth headers.
var ErrMCPUnauthorized = errors.New("mcp server rejected the credentials")

type McpRequesterInterface interface {
	ListTools(ctx context.Context, mcpConnection *models.MCPConnection) ([]*models.Tool, error)
	CallTool(
		ctx context.Context,
		mcpConnection *models.MCPConnection,
		tool *models.Tool,
		toolInput map[string]any,
	) (map[string]any, error)
}
