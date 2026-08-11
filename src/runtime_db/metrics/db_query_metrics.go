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

const meterName = "agentic_runtime_db"

type DbQueryMetrics struct {
	queryCounter      metric.Int64Counter
	durationHistogram metric.Float64Histogram
	rowsHistogram     metric.Int64Histogram
	errorCounter      metric.Int64Counter
}

func NewDbQueryMetrics() *DbQueryMetrics {
	noopMeter := noop.NewMeterProvider().Meter(meterName)
	meter := otel.GetMeterProvider().Meter(meterName)

	queryCounter, err := meter.Int64Counter(
		"db.query.count",
		metric.WithDescription("Total number of database queries"),
	)
	if err != nil {
		queryCounter, _ = noopMeter.Int64Counter("db.query.count")
	}

	durationHistogram, err := meter.Float64Histogram(
		"db.query.duration",
		metric.WithUnit("s"),
		metric.WithExplicitBucketBoundaries(common_metrics.LatencySecondsBuckets...),
		metric.WithDescription("Duration of database queries"),
	)
	if err != nil {
		durationHistogram, _ = noopMeter.Float64Histogram("db.query.duration")
	}

	rowsHistogram, err := meter.Int64Histogram(
		"db.query.rows",
		metric.WithDescription("Rows affected/returned by database queries"),
	)
	if err != nil {
		rowsHistogram, _ = noopMeter.Int64Histogram("db.query.rows")
	}

	errorCounter, err := meter.Int64Counter(
		"db.query.errors",
		metric.WithDescription("Total number of failed database queries"),
	)
	if err != nil {
		errorCounter, _ = noopMeter.Int64Counter("db.query.errors")
	}

	return &DbQueryMetrics{
		queryCounter:      queryCounter,
		durationHistogram: durationHistogram,
		rowsHistogram:     rowsHistogram,
		errorCounter:      errorCounter,
	}
}

// RecordQuery records one DB query: count + duration always, plus rows on success or an error on failure.
func (m *DbQueryMetrics) RecordQuery(
	ctx context.Context,
	operation, model string,
	duration time.Duration,
	rows int64,
	failed bool,
) {
	attrs := metric.WithAttributes(
		attribute.String("db.operation", operation),
		attribute.String("db.model", model),
	)
	m.queryCounter.Add(ctx, 1, attrs)
	m.durationHistogram.Record(ctx, duration.Seconds(), attrs)
	if failed {
		m.errorCounter.Add(ctx, 1, attrs)
		return
	}
	m.rowsHistogram.Record(ctx, rows, attrs)
}
