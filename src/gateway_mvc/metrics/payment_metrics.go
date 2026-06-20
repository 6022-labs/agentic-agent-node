package metrics

import (
	"context"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/metric"
	"go.opentelemetry.io/otel/metric/noop"
)

const paymentMeterName = "agentic_gateway_payment"

type PaymentMetrics struct {
	settlementCounter metric.Int64Counter
}

func NewPaymentMetrics() *PaymentMetrics {
	noopMeter := noop.NewMeterProvider().Meter(paymentMeterName)
	meter := otel.GetMeterProvider().Meter(paymentMeterName)

	settlementCounter, err := meter.Int64Counter(
		"payment.settlement",
		metric.WithDescription("Number of x402 payment settlements"),
	)
	if err != nil {
		settlementCounter, _ = noopMeter.Int64Counter("payment.settlement")
	}

	return &PaymentMetrics{settlementCounter: settlementCounter}
}

func (m *PaymentMetrics) RecordSettled(ctx context.Context, network string) {
	m.record(ctx, "settled", network)
}

func (m *PaymentMetrics) RecordRejected(ctx context.Context, network string) {
	m.record(ctx, "rejected", network)
}

func (m *PaymentMetrics) record(ctx context.Context, outcome, network string) {
	m.settlementCounter.Add(ctx, 1, metric.WithAttributes(
		attribute.String("outcome", outcome),
		attribute.String("network", network),
	))
}
