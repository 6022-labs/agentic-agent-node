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

const meterName = "agentic_runtime_vault"

type VaultMetrics struct {
	operationCounter  metric.Int64Counter
	durationHistogram metric.Float64Histogram
	errorCounter      metric.Int64Counter
}

func NewVaultMetrics() *VaultMetrics {
	noopMeter := noop.NewMeterProvider().Meter(meterName)
	meter := otel.GetMeterProvider().Meter(meterName)

	operationCounter, err := meter.Int64Counter(
		"vault.operation.count",
		metric.WithDescription("Total number of Vault secret operations"),
	)
	if err != nil {
		operationCounter, _ = noopMeter.Int64Counter("vault.operation.count")
	}

	durationHistogram, err := meter.Float64Histogram(
		"vault.operation.duration",
		metric.WithUnit("s"),
		metric.WithExplicitBucketBoundaries(common_metrics.LatencySecondsBuckets...),
		metric.WithDescription("Duration of Vault secret operations"),
	)
	if err != nil {
		durationHistogram, _ = noopMeter.Float64Histogram("vault.operation.duration")
	}

	errorCounter, err := meter.Int64Counter(
		"vault.operation.errors",
		metric.WithDescription("Total number of failed Vault secret operations"),
	)
	if err != nil {
		errorCounter, _ = noopMeter.Int64Counter("vault.operation.errors")
	}

	return &VaultMetrics{
		operationCounter:  operationCounter,
		durationHistogram: durationHistogram,
		errorCounter:      errorCounter,
	}
}

// RecordOperation records one Vault operation: count, duration, and (when failed) an error.
func (m *VaultMetrics) RecordOperation(ctx context.Context, operation, mount string, duration time.Duration, failed bool) {
	attrs := metric.WithAttributes(
		attribute.String("vault.operation", operation),
		attribute.String("vault.mount", mount),
	)
	m.operationCounter.Add(ctx, 1, attrs)
	m.durationHistogram.Record(ctx, duration.Seconds(), attrs)
	if failed {
		m.errorCounter.Add(ctx, 1, attrs)
	}
}
