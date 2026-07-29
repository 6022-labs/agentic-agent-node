package metrics

import (
	"context"
	"errors"
	"strconv"
	"time"

	"github.com/6022-labs/agentic-agent-node/src/agent_reasoning/models"
	"github.com/6022-labs/agentic-agent-node/src/agent_reasoning_http_mcp/clients"
	external_models "github.com/6022-labs/agentic-agent-node/src/agent_reasoning_http_mcp/models"
	common_metrics "github.com/6022-labs/agentic-agent-node/src/common/metrics"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/metric"
	"go.opentelemetry.io/otel/metric/noop"
)

const meterName = "agentic_runtime_http_mcp"

// Advisory bucket boundaries from the OTel MCP metrics semantic conventions.
var mcpDurationBuckets = []float64{0.01, 0.02, 0.05, 0.1, 0.2, 0.5, 1, 2, 5, 10, 30, 60, 120, 300}

type McpMetrics struct {
	operationDuration metric.Float64Histogram
}

func NewMcpMetrics() *McpMetrics {
	noopMeter := noop.NewMeterProvider().Meter(meterName)
	meter := otel.GetMeterProvider().Meter(meterName)

	operationDuration, err := meter.Float64Histogram(
		"mcp.client.operation.duration",
		metric.WithUnit("s"),
		metric.WithExplicitBucketBoundaries(mcpDurationBuckets...),
		metric.WithDescription("Duration of an MCP request, from send until the response is received"),
	)
	if err != nil {
		operationDuration, _ = noopMeter.Float64Histogram("mcp.client.operation.duration")
	}

	return &McpMetrics{operationDuration: operationDuration}
}

func (m *McpMetrics) RecordToolCall(
	ctx context.Context,
	mcpConnection *models.MCPConnection,
	toolName string,
	duration time.Duration,
	err error,
	rpcError *external_models.ExternalJsonRPCError,
	toolReportedError bool,
) {
	attrs := m.baseAttributes(clients.MethodCallTool, mcpConnection)
	attrs = append(attrs,
		attribute.String("gen_ai.tool.name", toolName),
		attribute.String("gen_ai.operation.name", "execute_tool"),
	)

	// A tool that reports its own failure completes the json-rpc call.
	if err == nil && rpcError == nil && toolReportedError {
		attrs = append(attrs, attribute.String("error.type", "tool_error"))
		m.operationDuration.Record(ctx, duration.Seconds(), metric.WithAttributes(attrs...))
		return
	}

	m.record(ctx, attrs, duration, err, rpcError)
}

func (m *McpMetrics) RecordListTools(
	ctx context.Context,
	mcpConnection *models.MCPConnection,
	duration time.Duration,
	err error,
	rpcError *external_models.ExternalJsonRPCError,
) {
	m.record(ctx, m.baseAttributes(clients.MethodListTools, mcpConnection), duration, err, rpcError)
}

func (m *McpMetrics) record(
	ctx context.Context,
	attrs []attribute.KeyValue,
	duration time.Duration,
	err error,
	rpcError *external_models.ExternalJsonRPCError,
) {
	attrs = append(attrs, m.failureAttributes(err, rpcError)...)
	m.operationDuration.Record(ctx, duration.Seconds(), metric.WithAttributes(attrs...))
}

func (m *McpMetrics) baseAttributes(method string, mcpConnection *models.MCPConnection) []attribute.KeyValue {
	attrs := []attribute.KeyValue{
		attribute.String("mcp.method.name", method),
		attribute.String("network.transport", "tcp"),
		attribute.String("network.protocol.name", "http"),
	}
	if mcpConnection == nil {
		return attrs
	}
	attrs = append(attrs, attribute.String("reasoning.mcp_connection", mcpConnection.Name))
	return append(attrs, common_metrics.ServerAttributes(mcpConnection.Url)...)
}

// failureAttributes stays empty on success: semconv forbids error.type on a completed operation.
func (m *McpMetrics) failureAttributes(
	err error,
	rpcError *external_models.ExternalJsonRPCError,
) []attribute.KeyValue {
	if err == nil && rpcError == nil {
		return nil
	}

	if rpcError != nil {
		code := strconv.Itoa(rpcError.Code)

		return []attribute.KeyValue{
			attribute.String("error.type", code),
			attribute.String("rpc.response.status_code", code),
		}
	}

	var statusErr *clients.McpStatusError
	if errors.As(err, &statusErr) {
		return []attribute.KeyValue{attribute.String("error.type", strconv.Itoa(statusErr.StatusCode))}
	}
	return []attribute.KeyValue{attribute.String("error.type", "_OTHER")}
}
