package metrics

import (
	"context"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/metric"
	"go.opentelemetry.io/otel/metric/noop"
)

type MintMetrics struct {
	counter metric.Int64Counter
}

func NewMintMetrics() *MintMetrics {
	counter, err := otel.GetMeterProvider().Meter(meterName).Int64Counter(
		"identity.mint",
		metric.WithDescription("Number of agent NFT mints reconciled"),
	)
	if err != nil {
		counter, _ = noop.NewMeterProvider().Meter(meterName).Int64Counter("identity.mint")
	}
	return &MintMetrics{counter: counter}
}

func (m *MintMetrics) RecordMintSucceeded(ctx context.Context) {
	m.counter.Add(ctx, 1, metric.WithAttributes(attribute.String("outcome", "success")))
}

func (m *MintMetrics) RecordMintFailed(ctx context.Context) {
	m.counter.Add(ctx, 1, metric.WithAttributes(attribute.String("outcome", "failure")))
}
