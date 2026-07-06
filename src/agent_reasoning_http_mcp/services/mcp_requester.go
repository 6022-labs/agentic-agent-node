package services

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/6022-labs/agentic-agent-node/src/agent_reasoning/models"
	"github.com/6022-labs/agentic-agent-node/src/agent_reasoning_http_mcp/clients"
	metrics_interfaces "github.com/6022-labs/agentic-agent-node/src/agent_reasoning_http_mcp/metrics/interfaces"
)

type MCPRequester struct {
	client  clients.McpClientInterface
	metrics metrics_interfaces.McpMetricsInterface
}

func NewMCPRequester(client clients.McpClientInterface, metrics metrics_interfaces.McpMetricsInterface) *MCPRequester {
	return &MCPRequester{
		client:  client,
		metrics: metrics,
	}
}

func (r *MCPRequester) ListTools(ctx context.Context, mcpConnection *models.MCPConnection) ([]*models.Tool, error) {
	if mcpConnection == nil {
		return nil, fmt.Errorf("MCP connection is nil")
	}

	start := time.Now()

	result, err := r.client.ListTools(ctx, mcpConnection.Url, mcpConnection.HeaderAuth)

	if err != nil {
		r.metrics.RecordListTools(ctx, mcpConnection.Name, time.Since(start), true)
		return nil, fmt.Errorf("failed to list MCP tools: %w", err)
	}
	if result.Error != nil {
		r.metrics.RecordListTools(ctx, mcpConnection.Name, time.Since(start), true)
		return nil, fmt.Errorf("MCP list tools returned error: %s", result.Error.Message)
	}
	r.metrics.RecordListTools(ctx, mcpConnection.Name, time.Since(start), false)

	return result.Result.ToFunctionTools(), nil
}

func (r *MCPRequester) CallTool(
	ctx context.Context,
	mcpConnection *models.MCPConnection,
	tool *models.Tool,
	toolInput map[string]any,
) (map[string]any, error) {
	if mcpConnection == nil {
		return nil, fmt.Errorf("MCP connection is nil")
	}
	if tool == nil {
		return nil, fmt.Errorf("MCP tool is nil")
	}

	fixedInput := r.fixToolInputBySchema(tool, toolInput)

	start := time.Now()

	result, err := r.client.CallTool(ctx, mcpConnection.Url, tool.Name, fixedInput, mcpConnection.HeaderAuth)

	if err != nil {
		r.metrics.RecordToolCall(ctx, mcpConnection.Name, tool.Name, time.Since(start), true)
		return nil, fmt.Errorf("failed to call MCP tool '%s': %w", tool.Name, err)
	}
	if result.Error != nil {
		r.metrics.RecordToolCall(ctx, mcpConnection.Name, tool.Name, time.Since(start), true)
		return nil, fmt.Errorf("MCP call tool '%s' returned error: %s", tool.Name, result.Error.Message)
	}
	r.metrics.RecordToolCall(ctx, mcpConnection.Name, tool.Name, time.Since(start), false)

	var output map[string]any
	if err := json.Unmarshal(result.Result, &output); err != nil {
		return nil, fmt.Errorf("failed to unmarshal MCP tool '%s' output: %w", tool.Name, err)
	}

	return output, nil
}

func (r *MCPRequester) fixToolInputBySchema(tool *models.Tool, input map[string]any) map[string]any {
	if input == nil {
		return nil
	}

	if len(tool.InputScheme.Properties) == 0 {
		return input
	}

	// Work on a shallow copy to avoid mutating the caller's map
	fixed := make(map[string]any, len(input))
	for k, v := range input {
		fixed[k] = v
	}

	for name, prop := range tool.InputScheme.Properties {
		val, ok := fixed[name]
		if !ok {
			continue
		}

		switch prop.Type {
		case "string":
			// If schema expects string but value is number/bool/[]byte, coerce to string
			switch typedVal := val.(type) {
			case float64, float32, int, int8, int16, int32, int64, uint, uint8, uint16, uint32, uint64, bool:
				fixed[name] = fmt.Sprint(typedVal)
			default:
				// no-op
			}
		default:
			// Additional coercions can be added here if needed (e.g. numeric strings to numbers).
		}
	}

	return fixed
}
