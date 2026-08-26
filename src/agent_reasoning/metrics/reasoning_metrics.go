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

const reasoningMeterName = "agentic_runtime_reasoning"

type ReasoningMetrics struct {
	loopDurationHistogram   metric.Float64Histogram
	loopIterationsHistogram metric.Int64Histogram
	toolCallCounter         metric.Int64Counter
	maxIterationsCounter    metric.Int64Counter
	abortedCounter          metric.Int64Counter
	conversationTitledCount metric.Int64Counter
}

func NewReasoningMetrics() *ReasoningMetrics {
	noopMeter := noop.NewMeterProvider().Meter(reasoningMeterName)
	meter := otel.GetMeterProvider().Meter(reasoningMeterName)

	loopDurationHistogram, err := meter.Float64Histogram(
		"reasoning.loop.duration",
		metric.WithUnit("s"),
		metric.WithExplicitBucketBoundaries(common_metrics.LongLatencySecondsBuckets...),
		metric.WithDescription("Duration of a full inference loop run"),
	)
	if err != nil {
		loopDurationHistogram, _ = noopMeter.Float64Histogram("reasoning.loop.duration")
	}

	loopIterationsHistogram, err := meter.Int64Histogram(
		"reasoning.loop.iterations",
		metric.WithUnit("{iteration}"),
		metric.WithExplicitBucketBoundaries(common_metrics.LoopIterationsBuckets...),
		metric.WithDescription("Number of LLM<->MCP iterations per inference loop run"),
	)
	if err != nil {
		loopIterationsHistogram, _ = noopMeter.Int64Histogram("reasoning.loop.iterations")
	}

	toolCallCounter, err := meter.Int64Counter(
		"reasoning.tool_calls",
		metric.WithUnit("{tool_call}"),
		metric.WithDescription("Number of tool calls dispatched during inference loops, self and MCP alike"),
	)
	if err != nil {
		toolCallCounter, _ = noopMeter.Int64Counter("reasoning.tool_calls")
	}

	maxIterationsCounter, err := meter.Int64Counter(
		"reasoning.loop.max_iterations_exceeded",
		metric.WithUnit("{run}"),
		metric.WithDescription("Number of inference loop runs that hit the iteration ceiling"),
	)
	if err != nil {
		maxIterationsCounter, _ = noopMeter.Int64Counter("reasoning.loop.max_iterations_exceeded")
	}

	abortedCounter, err := meter.Int64Counter(
		"reasoning.loop.aborted",
		metric.WithUnit("{run}"),
		metric.WithDescription("Number of inference loop runs aborted by a non-iteration safety ceiling"),
	)
	if err != nil {
		abortedCounter, _ = noopMeter.Int64Counter("reasoning.loop.aborted")
	}

	conversationTitledCount, err := meter.Int64Counter(
		"reasoning.conversation.titled",
		metric.WithUnit("{attempt}"),
		metric.WithDescription("Number of background conversation-title attempts by outcome"),
	)
	if err != nil {
		conversationTitledCount, _ = noopMeter.Int64Counter("reasoning.conversation.titled")
	}

	return &ReasoningMetrics{
		loopDurationHistogram:   loopDurationHistogram,
		loopIterationsHistogram: loopIterationsHistogram,
		toolCallCounter:         toolCallCounter,
		maxIterationsCounter:    maxIterationsCounter,
		abortedCounter:          abortedCounter,
		conversationTitledCount: conversationTitledCount,
	}
}

func (m *ReasoningMetrics) recordLoop(ctx context.Context, outcome string, duration time.Duration, iterations int64) {
	attrs := metric.WithAttributes(attribute.String("reasoning.outcome", outcome))
	m.loopDurationHistogram.Record(ctx, duration.Seconds(), attrs)
	m.loopIterationsHistogram.Record(ctx, iterations, attrs)
}

// RecordLoopCompleted records a loop that finished with a final answer.
func (m *ReasoningMetrics) RecordLoopCompleted(ctx context.Context, duration time.Duration, iterations int64) {
	m.recordLoop(ctx, "completed", duration, iterations)
}

// RecordLoopExceeded records a loop that hit the iteration ceiling.
func (m *ReasoningMetrics) RecordLoopExceeded(ctx context.Context, duration time.Duration, iterations int64) {
	m.recordLoop(ctx, "max_iterations_exceeded", duration, iterations)
	m.maxIterationsCounter.Add(ctx, 1)
}

func (m *ReasoningMetrics) RecordLoopAborted(
	ctx context.Context,
	reason string,
	duration time.Duration,
	iterations int64,
) {
	m.recordLoop(ctx, reason, duration, iterations)
	m.abortedCounter.Add(ctx, 1, metric.WithAttributes(attribute.String("reasoning.abort_reason", reason)))
}

func (m *ReasoningMetrics) RecordToolCalls(ctx context.Context, count int64) {
	m.toolCallCounter.Add(ctx, count)
}

func (m *ReasoningMetrics) RecordConversationTitled(ctx context.Context, outcome string) {
	m.conversationTitledCount.Add(ctx, 1, metric.WithAttributes(attribute.String("reasoning.outcome", outcome)))
}
