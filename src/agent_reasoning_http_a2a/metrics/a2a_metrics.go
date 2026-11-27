package metrics

import (
	"context"
	"net/url"
	"time"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/metric"
	"go.opentelemetry.io/otel/metric/noop"
)

const meterName = "agentic_runtime_http_a2a"

var a2aDurationBuckets = []float64{0.01, 0.05, 0.1, 0.5, 1, 2, 5, 10, 30, 60, 120, 300}

type A2AMetrics struct {
	operationDuration metric.Float64Histogram
}

func NewA2AMetrics() *A2AMetrics {
	noopMeter := noop.NewMeterProvider().Meter(meterName)
	meter := otel.GetMeterProvider().Meter(meterName)

	operationDuration, err := meter.Float64Histogram(
		"a2a.client.operation.duration",
		metric.WithUnit("s"),
		metric.WithExplicitBucketBoundaries(a2aDurationBuckets...),
		metric.WithDescription("Duration of an outbound A2A request, from send until the response is received"),
	)
	if err != nil {
		operationDuration, _ = noopMeter.Float64Histogram("a2a.client.operation.duration")
	}

	return &A2AMetrics{operationDuration: operationDuration}
}

func (m *A2AMetrics) RecordSendMessage(
	ctx context.Context,
	endpointUrl string,
	duration time.Duration,
	err error,
	paymentRequired bool,
) {
	attrs := m.baseAttributes("send_message", endpointUrl)
	if paymentRequired {
		attrs = append(attrs, attribute.String("error.type", "payment_required"))
	} else if err != nil {
		attrs = append(attrs, attribute.String("error.type", "request_error"))
	}
	m.operationDuration.Record(ctx, duration.Seconds(), metric.WithAttributes(attrs...))
}

func (m *A2AMetrics) RecordFetchAgentCard(ctx context.Context, baseUrl string, duration time.Duration, err error) {
	attrs := m.baseAttributes("fetch_agent_card", baseUrl)
	if err != nil {
		attrs = append(attrs, attribute.String("error.type", "request_error"))
	}
	m.operationDuration.Record(ctx, duration.Seconds(), metric.WithAttributes(attrs...))
}

// baseAttributes keys by server, never by full url: a per-agent path would explode cardinality.
func (m *A2AMetrics) baseAttributes(operation string, rawUrl string) []attribute.KeyValue {
	attrs := []attribute.KeyValue{attribute.String("a2a.operation.name", operation)}
	if parsed, err := url.Parse(rawUrl); err == nil && len(parsed.Host) != 0 {
		attrs = append(attrs, attribute.String("server.address", parsed.Host))
	}
	return attrs
}
