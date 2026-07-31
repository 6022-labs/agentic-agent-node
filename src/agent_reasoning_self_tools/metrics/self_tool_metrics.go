package metrics

import (
	"context"
	"errors"
	"time"

	apperrors "github.com/6022-labs/agentic-agent-node/src/common/errors"
	common_metrics "github.com/6022-labs/agentic-agent-node/src/common/metrics"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/metric"
	"go.opentelemetry.io/otel/metric/noop"
)

const meterName = "agentic_runtime_self_tools"

type SelfToolMetrics struct {
	executeToolDuration metric.Float64Histogram
}

func NewSelfToolMetrics() *SelfToolMetrics {
	noopMeter := noop.NewMeterProvider().Meter(meterName)
	meter := otel.GetMeterProvider().Meter(meterName)

	executeToolDuration, err := meter.Float64Histogram(
		"gen_ai.execute_tool.duration",
		metric.WithUnit("s"),
		metric.WithExplicitBucketBoundaries(common_metrics.GenAiDurationSecondsBuckets...),
		metric.WithDescription("The duration of a single tool execution"),
	)
	if err != nil {
		executeToolDuration, _ = noopMeter.Float64Histogram("gen_ai.execute_tool.duration")
	}

	return &SelfToolMetrics{executeToolDuration: executeToolDuration}
}

func (m *SelfToolMetrics) RecordToolCall(ctx context.Context, toolName string, duration time.Duration, err error) {
	attrs := []attribute.KeyValue{
		attribute.String("gen_ai.tool.name", toolName),
		// Self tools are predefined functions the model supplies parameters for.
		attribute.String("gen_ai.tool.type", "function"),
	}
	attrs = append(attrs, m.failureAttributes(err)...)

	m.executeToolDuration.Record(ctx, duration.Seconds(), metric.WithAttributes(attrs...))
}

// failureAttributes stays empty on success: semconv forbids error.type on a completed operation.
func (m *SelfToolMetrics) failureAttributes(err error) []attribute.KeyValue {
	if err == nil {
		return nil
	}
	return []attribute.KeyValue{
		attribute.String("error.type", m.errorType(err)),
	}
}

// errorType keeps cardinality low by reporting the app error's kind, not its message.
func (*SelfToolMetrics) errorType(err error) string {
	var validationError *apperrors.ValidationError
	if errors.As(err, &validationError) {
		return "validation"
	}
	var notFoundError *apperrors.NotFoundError
	if errors.As(err, &notFoundError) {
		return "not_found"
	}
	var conflictError *apperrors.ConflictError
	if errors.As(err, &conflictError) {
		return "conflict"
	}
	var unauthorizedError *apperrors.UnauthorizedError
	if errors.As(err, &unauthorizedError) {
		return "unauthorized"
	}
	var unavailableError *apperrors.UnavailableError
	if errors.As(err, &unavailableError) {
		return "unavailable"
	}
	var internalError *apperrors.InternalError
	if errors.As(err, &internalError) {
		return "internal"
	}
	return "_OTHER"
}
