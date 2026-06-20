package metrics

import (
	"context"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/metric"
	"go.opentelemetry.io/otel/metric/noop"
)

type RegistrationMetrics struct {
	counter metric.Int64Counter
}

func NewRegistrationMetrics() *RegistrationMetrics {
	counter, err := otel.GetMeterProvider().Meter(meterName).Int64Counter(
		"identity.registration",
		metric.WithDescription("Number of agent self-registration attempts"),
	)
	if err != nil {
		counter, _ = noop.NewMeterProvider().Meter(meterName).Int64Counter("identity.registration")
	}
	return &RegistrationMetrics{counter: counter}
}

func (m *RegistrationMetrics) RecordRegistrationSucceeded(ctx context.Context) {
	m.counter.Add(ctx, 1, metric.WithAttributes(attribute.String("outcome", "success")))
}

func (m *RegistrationMetrics) RecordRegistrationPending(ctx context.Context) {
	m.counter.Add(ctx, 1, metric.WithAttributes(attribute.String("outcome", "pending")))
}

func (m *RegistrationMetrics) RecordRegistrationFailed(ctx context.Context) {
	m.counter.Add(ctx, 1, metric.WithAttributes(attribute.String("outcome", "failure")))
}
