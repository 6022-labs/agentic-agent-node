package metrics

import (
	"context"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/metric"
	"go.opentelemetry.io/otel/metric/noop"
)

const authMeterName = "agentic_gateway_auth"

type AuthMetrics struct {
	successCounter metric.Int64Counter
	failureCounter metric.Int64Counter
}

func NewAuthMetrics() *AuthMetrics {
	noopMeter := noop.NewMeterProvider().Meter(authMeterName)
	meter := otel.GetMeterProvider().Meter(authMeterName)

	successCounter, err := meter.Int64Counter(
		"gateway.auth.success",
		metric.WithDescription("Number of successful wallet auth verifications (JWT issued)"),
	)
	if err != nil {
		successCounter, _ = noopMeter.Int64Counter("gateway.auth.success")
	}

	failureCounter, err := meter.Int64Counter(
		"gateway.auth.failure",
		metric.WithDescription("Number of failed wallet auth verifications"),
	)
	if err != nil {
		failureCounter, _ = noopMeter.Int64Counter("gateway.auth.failure")
	}

	return &AuthMetrics{
		successCounter: successCounter,
		failureCounter: failureCounter,
	}
}

func (m *AuthMetrics) RecordAuthSuccess(ctx context.Context) {
	m.successCounter.Add(ctx, 1)
}

func (m *AuthMetrics) RecordAuthFailure(ctx context.Context) {
	m.failureCounter.Add(ctx, 1)
}
