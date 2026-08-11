package interfaces

import (
	"context"
	"time"

	"github.com/6022-labs/agentic-agent-node/src/agent_reasoning/models"
	external_models "github.com/6022-labs/agentic-agent-node/src/agent_reasoning_http_mcp/models"
)

// McpMetricsInterface records the OTel MCP-semconv client metric; callers pass the raw outcome, not semconv values.
type McpMetricsInterface interface {
	// RecordToolCall records one tools/call; err and rpcError are nil on success.
	RecordToolCall(
		ctx context.Context,
		mcpConnection *models.MCPConnection,
		toolName string,
		duration time.Duration,
		err error,
		rpcError *external_models.ExternalJsonRPCError,
		toolReportedError bool,
	)

	// RecordListTools records one tools/list; err and rpcError are nil on success.
	RecordListTools(
		ctx context.Context,
		mcpConnection *models.MCPConnection,
		duration time.Duration,
		err error,
		rpcError *external_models.ExternalJsonRPCError,
	)
}
