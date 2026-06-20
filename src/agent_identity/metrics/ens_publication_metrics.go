package metrics

import (
	"context"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/metric"
	"go.opentelemetry.io/otel/metric/noop"
)

type EnsPublicationMetrics struct {
	counter metric.Int64Counter
}

func NewEnsPublicationMetrics() *EnsPublicationMetrics {
	counter, err := otel.GetMeterProvider().Meter(meterName).Int64Counter(
		"identity.ens.published",
		metric.WithDescription("Number of ENS runtime-record publications"),
	)
	if err != nil {
		counter, _ = noop.NewMeterProvider().Meter(meterName).Int64Counter("identity.ens.published")
	}
	return &EnsPublicationMetrics{counter: counter}
}

func (m *EnsPublicationMetrics) RecordEnsPublishedSucceeded(ctx context.Context) {
	m.counter.Add(ctx, 1, metric.WithAttributes(attribute.String("outcome", "success")))
}

func (m *EnsPublicationMetrics) RecordEnsPublishedFailed(ctx context.Context) {
	m.counter.Add(ctx, 1, metric.WithAttributes(attribute.String("outcome", "failure")))
}

func (m *EnsPublicationMetrics) RecordEnsPublishedNotMinted(ctx context.Context) {
	m.counter.Add(ctx, 1, metric.WithAttributes(attribute.String("outcome", "not_minted")))
}
