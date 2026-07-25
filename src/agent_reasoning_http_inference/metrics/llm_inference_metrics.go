package metrics

import (
	"context"
	"time"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/metric"
	"go.opentelemetry.io/otel/metric/noop"
)

const meterName = "agentic_runtime_http_inference"

// Advisory bucket boundaries from the OTel GenAI metrics semantic conventions.
var (
	genAiDurationBuckets = []float64{0.01, 0.02, 0.04, 0.08, 0.16, 0.32, 0.64, 1.28, 2.56, 5.12, 10.24, 20.48, 40.96, 81.92}
	genAiTokenBuckets    = []float64{1, 4, 16, 64, 256, 1024, 4096, 16384, 65536, 262144, 1048576, 4194304, 16777216, 67108864}
)

type LlmInferenceMetrics struct {
	durationHistogram   metric.Float64Histogram
	tokenUsageHistogram metric.Int64Histogram
}

func NewLlmInferenceMetrics() *LlmInferenceMetrics {
	noopMeter := noop.NewMeterProvider().Meter(meterName)
	meter := otel.GetMeterProvider().Meter(meterName)

	durationHistogram, err := meter.Float64Histogram(
		"gen_ai.client.operation.duration",
		metric.WithUnit("s"),
		metric.WithExplicitBucketBoundaries(genAiDurationBuckets...),
		metric.WithDescription("Duration of LLM chat-completion calls"),
	)
	if err != nil {
		durationHistogram, _ = noopMeter.Float64Histogram("gen_ai.client.operation.duration")
	}

	tokenUsageHistogram, err := meter.Int64Histogram(
		"gen_ai.client.token.usage",
		metric.WithUnit("{token}"),
		metric.WithExplicitBucketBoundaries(genAiTokenBuckets...),
		metric.WithDescription("Number of tokens used in LLM chat-completion calls"),
	)
	if err != nil {
		tokenUsageHistogram, _ = noopMeter.Int64Histogram("gen_ai.client.token.usage")
	}

	return &LlmInferenceMetrics{
		durationHistogram:   durationHistogram,
		tokenUsageHistogram: tokenUsageHistogram,
	}
}

// RecordOperationDuration records a chat-completion call. errorType is the
// gen_ai error.type attribute value, or "" when the call succeeded.
func (m *LlmInferenceMetrics) RecordOperationDuration(ctx context.Context, provider, model string, duration time.Duration, errorType string) {
	attrs := []attribute.KeyValue{
		attribute.String("gen_ai.provider.name", provider),
		attribute.String("gen_ai.request.model", model),
		attribute.String("gen_ai.operation.name", "chat"),
	}
	if errorType != "" {
		attrs = append(attrs, attribute.String("error.type", errorType))
	}
	m.durationHistogram.Record(ctx, duration.Seconds(), metric.WithAttributes(attrs...))
}

// RecordProbe records a connection-validation probe on the operation duration histogram.
func (m *LlmInferenceMetrics) RecordProbe(ctx context.Context, provider, model string, duration time.Duration, failed bool) {
	attrs := []attribute.KeyValue{
		attribute.String("gen_ai.provider.name", provider),
		attribute.String("gen_ai.request.model", model),
		attribute.String("gen_ai.operation.name", "probe"),
	}
	if failed {
		attrs = append(attrs, attribute.String("error.type", "_OTHER"))
	}
	m.durationHistogram.Record(ctx, duration.Seconds(), metric.WithAttributes(attrs...))
}

// RecordTokenUsage records token usage for a chat-completion call. tokenType is
// the gen_ai.token.type value ("input" / "output").
func (m *LlmInferenceMetrics) RecordTokenUsage(ctx context.Context, provider, model, tokenType string, tokens int64) {
	m.tokenUsageHistogram.Record(ctx, tokens, metric.WithAttributes(
		attribute.String("gen_ai.provider.name", provider),
		attribute.String("gen_ai.request.model", model),
		attribute.String("gen_ai.operation.name", "chat"),
		attribute.String("gen_ai.token.type", tokenType),
	))
}
