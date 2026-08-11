package metrics

import (
	"context"
	"errors"
	"strconv"
	"strings"
	"time"

	"github.com/6022-labs/agentic-agent-node/src/agent_reasoning_http_inference/clients"

	"github.com/6022-labs/agentic-agent-node/src/agent_reasoning_http_inference/metrics/interfaces"
	common_metrics "github.com/6022-labs/agentic-agent-node/src/common/metrics"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/metric"
	"go.opentelemetry.io/otel/metric/noop"
)

const meterName = "agentic_runtime_http_inference"

// callKindKey is ours, not semconv: a probe reports operation.name "chat", so it needs its own discriminator.
const callKindKey = "reasoning.call_kind"

// Advisory bucket boundaries from the OTel GenAI metrics semantic conventions.
var (
	genAiTokenBuckets = []float64{
		1,
		4,
		16,
		64,
		256,
		1024,
		4096,
		16384,
		65536,
		262144,
		1048576,
		4194304,
		16777216,
		67108864,
	}
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
		metric.WithExplicitBucketBoundaries(common_metrics.GenAiDurationSecondsBuckets...),
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

func (m *LlmInferenceMetrics) RecordOperationDuration(
	ctx context.Context,
	call interfaces.CompletionCall,
	duration time.Duration,
	err error,
) {
	attrs := append(m.callAttributes(call), attribute.String(callKindKey, "inference"))
	attrs = append(attrs, m.failureAttributes(err)...)
	m.durationHistogram.Record(ctx, duration.Seconds(), metric.WithAttributes(attrs...))
}

func (m *LlmInferenceMetrics) RecordProbe(
	ctx context.Context,
	call interfaces.CompletionCall,
	duration time.Duration,
	err error,
) {
	attrs := append(m.callAttributes(call), attribute.String(callKindKey, "probe"))
	attrs = append(attrs, m.failureAttributes(err)...)
	m.durationHistogram.Record(ctx, duration.Seconds(), metric.WithAttributes(attrs...))
}

func (m *LlmInferenceMetrics) RecordTokenUsage(
	ctx context.Context,
	call interfaces.CompletionCall,
	inputTokens, outputTokens int64,
) {
	m.recordTokens(ctx, call, "input", inputTokens)
	m.recordTokens(ctx, call, "output", outputTokens)
}

func (m *LlmInferenceMetrics) recordTokens(
	ctx context.Context,
	call interfaces.CompletionCall,
	tokenType string,
	tokens int64,
) {
	attrs := append(m.callAttributes(call), attribute.String("gen_ai.token.type", tokenType))
	m.tokenUsageHistogram.Record(ctx, tokens, metric.WithAttributes(attrs...))
}

// failureAttributes stays empty on success: semconv forbids error.type on a completed operation.
func (m *LlmInferenceMetrics) failureAttributes(err error) []attribute.KeyValue {
	if err == nil {
		return nil
	}

	var statusErr *clients.LlmStatusError
	if errors.As(err, &statusErr) {
		return []attribute.KeyValue{attribute.String("error.type", strconv.Itoa(statusErr.StatusCode))}
	}
	return []attribute.KeyValue{attribute.String("error.type", "_OTHER")}
}

// callAttributes builds a fresh slice every call; a shared one would let concurrent records clobber each other.
func (m *LlmInferenceMetrics) callAttributes(call interfaces.CompletionCall) []attribute.KeyValue {
	attrs := make([]attribute.KeyValue, 0, 5)
	attrs = append(attrs,
		attribute.String("gen_ai.provider.name", call.Provider),
		attribute.String("gen_ai.request.model", call.RequestModel),
		attribute.String("gen_ai.operation.name", "chat"),
	)
	if len(strings.TrimSpace(call.ResponseModel)) != 0 {
		attrs = append(attrs, attribute.String("gen_ai.response.model", call.ResponseModel))
	}
	return append(attrs, common_metrics.ServerAttributes(call.ServerUrl)...)
}
