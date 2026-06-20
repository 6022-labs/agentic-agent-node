package metrics

import (
	"context"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/metric"
	"go.opentelemetry.io/otel/metric/noop"
)

const meterName = "agentic_runtime_identity"

type BootstrapMetrics struct {
	counter metric.Int64Counter
}

func NewBootstrapMetrics() *BootstrapMetrics {
	counter, err := otel.GetMeterProvider().Meter(meterName).Int64Counter(
		"identity.bootstrap",
		metric.WithDescription("Number of agent bootstrap runs"),
	)
	if err != nil {
		counter, _ = noop.NewMeterProvider().Meter(meterName).Int64Counter("identity.bootstrap")
	}
	return &BootstrapMetrics{counter: counter}
}

func (m *BootstrapMetrics) RecordBootstrapSucceeded(ctx context.Context) {
	m.counter.Add(ctx, 1, metric.WithAttributes(attribute.String("outcome", "success")))
}

func (m *BootstrapMetrics) RecordBootstrapFailed(ctx context.Context) {
	m.counter.Add(ctx, 1, metric.WithAttributes(attribute.String("outcome", "failure")))
}
