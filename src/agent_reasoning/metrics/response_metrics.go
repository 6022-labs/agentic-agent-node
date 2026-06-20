package metrics

import (
	"context"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/metric"
	"go.opentelemetry.io/otel/metric/noop"
)

type ResponseMetrics struct {
	counter metric.Int64Counter
}

func NewResponseMetrics() *ResponseMetrics {
	counter, err := otel.GetMeterProvider().Meter(reasoningMeterName).Int64Counter(
		"reasoning.response",
		metric.WithDescription("Number of agent responses generated"),
	)
	if err != nil {
		counter, _ = noop.NewMeterProvider().Meter(reasoningMeterName).Int64Counter("reasoning.response")
	}
	return &ResponseMetrics{counter: counter}
}

func (m *ResponseMetrics) RecordResponseSucceeded(ctx context.Context) {
	m.counter.Add(ctx, 1, metric.WithAttributes(attribute.String("outcome", "success")))
}

func (m *ResponseMetrics) RecordResponseFailed(ctx context.Context) {
	m.counter.Add(ctx, 1, metric.WithAttributes(attribute.String("outcome", "failure")))
}
