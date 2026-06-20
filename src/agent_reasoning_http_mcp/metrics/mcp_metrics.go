package metrics

import (
	"context"
	"time"

	common_metrics "github.com/6022-labs/agentic-agent-node/src/common/metrics"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/metric"
	"go.opentelemetry.io/otel/metric/noop"
)

const meterName = "agentic_runtime_http_mcp"

type McpMetrics struct {
	toolCallCounter       metric.Int64Counter
	toolCallDuration      metric.Float64Histogram
	toolCallErrorCounter  metric.Int64Counter
	listToolsCounter      metric.Int64Counter
	listToolsDuration     metric.Float64Histogram
	listToolsErrorCounter metric.Int64Counter
}

func NewMcpMetrics() *McpMetrics {
	noopMeter := noop.NewMeterProvider().Meter(meterName)
	meter := otel.GetMeterProvider().Meter(meterName)

	toolCallCounter, err := meter.Int64Counter(
		"mcp.tool.call.count",
		metric.WithDescription("Total number of MCP tool calls"),
	)
	if err != nil {
		toolCallCounter, _ = noopMeter.Int64Counter("mcp.tool.call.count")
	}

	toolCallDuration, err := meter.Float64Histogram(
		"mcp.tool.call.duration",
		metric.WithUnit("s"),
		metric.WithExplicitBucketBoundaries(common_metrics.LatencySecondsBuckets...),
		metric.WithDescription("Duration of MCP tool calls"),
	)
	if err != nil {
		toolCallDuration, _ = noopMeter.Float64Histogram("mcp.tool.call.duration")
	}

	toolCallErrorCounter, err := meter.Int64Counter(
		"mcp.tool.call.errors",
		metric.WithDescription("Total number of MCP tool call errors"),
	)
	if err != nil {
		toolCallErrorCounter, _ = noopMeter.Int64Counter("mcp.tool.call.errors")
	}

	listToolsCounter, err := meter.Int64Counter(
		"mcp.list_tools.count",
		metric.WithDescription("Total number of MCP list-tools requests"),
	)
	if err != nil {
		listToolsCounter, _ = noopMeter.Int64Counter("mcp.list_tools.count")
	}

	listToolsDuration, err := meter.Float64Histogram(
		"mcp.list_tools.duration",
		metric.WithUnit("s"),
		metric.WithExplicitBucketBoundaries(common_metrics.LatencySecondsBuckets...),
		metric.WithDescription("Duration of MCP list-tools requests"),
	)
	if err != nil {
		listToolsDuration, _ = noopMeter.Float64Histogram("mcp.list_tools.duration")
	}

	listToolsErrorCounter, err := meter.Int64Counter(
		"mcp.list_tools.errors",
		metric.WithDescription("Total number of MCP list-tools errors"),
	)
	if err != nil {
		listToolsErrorCounter, _ = noopMeter.Int64Counter("mcp.list_tools.errors")
	}

	return &McpMetrics{
		toolCallCounter:       toolCallCounter,
		toolCallDuration:      toolCallDuration,
		toolCallErrorCounter:  toolCallErrorCounter,
		listToolsCounter:      listToolsCounter,
		listToolsDuration:     listToolsDuration,
		listToolsErrorCounter: listToolsErrorCounter,
	}
}

// RecordToolCall records one MCP tool call: count, duration, and (when failed) an error.
func (m *McpMetrics) RecordToolCall(ctx context.Context, server, tool string, duration time.Duration, failed bool) {
	attrs := metric.WithAttributes(
		attribute.String("mcp.server", server),
		attribute.String("mcp.tool.name", tool),
	)
	m.toolCallCounter.Add(ctx, 1, attrs)
	m.toolCallDuration.Record(ctx, duration.Seconds(), attrs)
	if failed {
		m.toolCallErrorCounter.Add(ctx, 1, attrs)
	}
}

// RecordListTools records one MCP list-tools request: count, duration, and (when failed) an error.
func (m *McpMetrics) RecordListTools(ctx context.Context, server string, duration time.Duration, failed bool) {
	attrs := metric.WithAttributes(attribute.String("mcp.server", server))
	m.listToolsCounter.Add(ctx, 1, attrs)
	m.listToolsDuration.Record(ctx, duration.Seconds(), attrs)
	if failed {
		m.listToolsErrorCounter.Add(ctx, 1, attrs)
	}
}
