package interfaces

import (
	"context"
	"time"
)

// McpMetricsInterface records MCP tool-call and list-tools health.
type McpMetricsInterface interface {
	RecordToolCall(ctx context.Context, server, tool string, duration time.Duration, failed bool)
	RecordListTools(ctx context.Context, server string, duration time.Duration, failed bool)
}
