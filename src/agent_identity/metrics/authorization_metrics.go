package metrics

import (
	"context"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/metric"
	"go.opentelemetry.io/otel/metric/noop"
)

type AuthorizationMetrics struct {
	counter metric.Int64Counter
}

func NewAuthorizationMetrics() *AuthorizationMetrics {
	counter, err := otel.GetMeterProvider().Meter(meterName).Int64Counter(
		"identity.authorization",
		metric.WithDescription("Number of agent authorization checks"),
	)
	if err != nil {
		counter, _ = noop.NewMeterProvider().Meter(meterName).Int64Counter("identity.authorization")
	}
	return &AuthorizationMetrics{counter: counter}
}

func (m *AuthorizationMetrics) RecordAuthorizationAuthorized(ctx context.Context) {
	m.counter.Add(ctx, 1, metric.WithAttributes(attribute.String("outcome", "authorized")))
}

func (m *AuthorizationMetrics) RecordAuthorizationDenied(ctx context.Context) {
	m.counter.Add(ctx, 1, metric.WithAttributes(attribute.String("outcome", "denied")))
}

func (m *AuthorizationMetrics) RecordAuthorizationFailed(ctx context.Context) {
	m.counter.Add(ctx, 1, metric.WithAttributes(attribute.String("outcome", "failure")))
}
