package observability

import (
	"context"

	metricapi "go.opentelemetry.io/otel/metric"
	"go.opentelemetry.io/otel/metric/embedded"
)

// NewAgentIdentityMeterProvider adds the ctx agent identity to sync measurements (observables pass through).
func NewAgentIdentityMeterProvider(inner metricapi.MeterProvider) metricapi.MeterProvider {
	return ctxMeterProvider{inner: inner}
}

func addAttrs(ctx context.Context, opts []metricapi.AddOption) []metricapi.AddOption {
	if identity, ok := AgentIdentityFromContext(ctx); ok {
		if attrs := identity.attributes(); len(attrs) != 0 {
			return append(opts, metricapi.WithAttributes(attrs...))
		}
	}
	return opts
}

func recordAttrs(ctx context.Context, opts []metricapi.RecordOption) []metricapi.RecordOption {
	if identity, ok := AgentIdentityFromContext(ctx); ok {
		if attrs := identity.attributes(); len(attrs) != 0 {
			return append(opts, metricapi.WithAttributes(attrs...))
		}
	}
	return opts
}

type ctxMeterProvider struct {
	embedded.MeterProvider
	inner metricapi.MeterProvider
}

func (p ctxMeterProvider) Meter(name string, opts ...metricapi.MeterOption) metricapi.Meter {
	return ctxMeter{inner: p.inner.Meter(name, opts...)}
}

type ctxMeter struct {
	embedded.Meter
	inner metricapi.Meter
}

func (m ctxMeter) Int64Counter(name string, opts ...metricapi.Int64CounterOption) (metricapi.Int64Counter, error) {
	inst, err := m.inner.Int64Counter(name, opts...)
	if err != nil {
		return nil, err
	}
	return ctxInt64Counter{Int64Counter: inst}, nil
}

func (m ctxMeter) Int64UpDownCounter(
	name string,
	opts ...metricapi.Int64UpDownCounterOption,
) (metricapi.Int64UpDownCounter, error) {
	inst, err := m.inner.Int64UpDownCounter(name, opts...)
	if err != nil {
		return nil, err
	}
	return ctxInt64UpDownCounter{Int64UpDownCounter: inst}, nil
}

func (m ctxMeter) Int64Histogram(
	name string,
	opts ...metricapi.Int64HistogramOption,
) (metricapi.Int64Histogram, error) {
	inst, err := m.inner.Int64Histogram(name, opts...)
	if err != nil {
		return nil, err
	}
	return ctxInt64Histogram{Int64Histogram: inst}, nil
}

func (m ctxMeter) Int64Gauge(name string, opts ...metricapi.Int64GaugeOption) (metricapi.Int64Gauge, error) {
	inst, err := m.inner.Int64Gauge(name, opts...)
	if err != nil {
		return nil, err
	}
	return ctxInt64Gauge{Int64Gauge: inst}, nil
}

func (m ctxMeter) Float64Counter(
	name string,
	opts ...metricapi.Float64CounterOption,
) (metricapi.Float64Counter, error) {
	inst, err := m.inner.Float64Counter(name, opts...)
	if err != nil {
		return nil, err
	}
	return ctxFloat64Counter{Float64Counter: inst}, nil
}

func (m ctxMeter) Float64UpDownCounter(
	name string,
	opts ...metricapi.Float64UpDownCounterOption,
) (metricapi.Float64UpDownCounter, error) {
	inst, err := m.inner.Float64UpDownCounter(name, opts...)
	if err != nil {
		return nil, err
	}
	return ctxFloat64UpDownCounter{Float64UpDownCounter: inst}, nil
}

func (m ctxMeter) Float64Histogram(
	name string,
	opts ...metricapi.Float64HistogramOption,
) (metricapi.Float64Histogram, error) {
	inst, err := m.inner.Float64Histogram(name, opts...)
	if err != nil {
		return nil, err
	}
	return ctxFloat64Histogram{Float64Histogram: inst}, nil
}

func (m ctxMeter) Float64Gauge(name string, opts ...metricapi.Float64GaugeOption) (metricapi.Float64Gauge, error) {
	inst, err := m.inner.Float64Gauge(name, opts...)
	if err != nil {
		return nil, err
	}
	return ctxFloat64Gauge{Float64Gauge: inst}, nil
}

// Observable instruments and callbacks delegate unchanged — their collection ctx carries no agent.
func (m ctxMeter) Int64ObservableCounter(
	name string,
	opts ...metricapi.Int64ObservableCounterOption,
) (metricapi.Int64ObservableCounter, error) {
	return m.inner.Int64ObservableCounter(name, opts...)
}

func (m ctxMeter) Int64ObservableUpDownCounter(
	name string,
	opts ...metricapi.Int64ObservableUpDownCounterOption,
) (metricapi.Int64ObservableUpDownCounter, error) {
	return m.inner.Int64ObservableUpDownCounter(name, opts...)
}

func (m ctxMeter) Int64ObservableGauge(
	name string,
	opts ...metricapi.Int64ObservableGaugeOption,
) (metricapi.Int64ObservableGauge, error) {
	return m.inner.Int64ObservableGauge(name, opts...)
}

func (m ctxMeter) Float64ObservableCounter(
	name string,
	opts ...metricapi.Float64ObservableCounterOption,
) (metricapi.Float64ObservableCounter, error) {
	return m.inner.Float64ObservableCounter(name, opts...)
}

func (m ctxMeter) Float64ObservableUpDownCounter(
	name string,
	opts ...metricapi.Float64ObservableUpDownCounterOption,
) (metricapi.Float64ObservableUpDownCounter, error) {
	return m.inner.Float64ObservableUpDownCounter(name, opts...)
}

func (m ctxMeter) Float64ObservableGauge(
	name string,
	opts ...metricapi.Float64ObservableGaugeOption,
) (metricapi.Float64ObservableGauge, error) {
	return m.inner.Float64ObservableGauge(name, opts...)
}

func (m ctxMeter) RegisterCallback(
	callback metricapi.Callback,
	instruments ...metricapi.Observable,
) (metricapi.Registration, error) {
	return m.inner.RegisterCallback(callback, instruments...)
}

type ctxInt64Counter struct {
	metricapi.Int64Counter
}

func (c ctxInt64Counter) Add(ctx context.Context, value int64, opts ...metricapi.AddOption) {
	c.Int64Counter.Add(ctx, value, addAttrs(ctx, opts)...)
}

type ctxInt64UpDownCounter struct {
	metricapi.Int64UpDownCounter
}

func (c ctxInt64UpDownCounter) Add(ctx context.Context, value int64, opts ...metricapi.AddOption) {
	c.Int64UpDownCounter.Add(ctx, value, addAttrs(ctx, opts)...)
}

type ctxInt64Histogram struct {
	metricapi.Int64Histogram
}

func (h ctxInt64Histogram) Record(ctx context.Context, value int64, opts ...metricapi.RecordOption) {
	h.Int64Histogram.Record(ctx, value, recordAttrs(ctx, opts)...)
}

type ctxInt64Gauge struct {
	metricapi.Int64Gauge
}

func (g ctxInt64Gauge) Record(ctx context.Context, value int64, opts ...metricapi.RecordOption) {
	g.Int64Gauge.Record(ctx, value, recordAttrs(ctx, opts)...)
}

type ctxFloat64Counter struct {
	metricapi.Float64Counter
}

func (c ctxFloat64Counter) Add(ctx context.Context, value float64, opts ...metricapi.AddOption) {
	c.Float64Counter.Add(ctx, value, addAttrs(ctx, opts)...)
}

type ctxFloat64UpDownCounter struct {
	metricapi.Float64UpDownCounter
}

func (c ctxFloat64UpDownCounter) Add(ctx context.Context, value float64, opts ...metricapi.AddOption) {
	c.Float64UpDownCounter.Add(ctx, value, addAttrs(ctx, opts)...)
}

type ctxFloat64Histogram struct {
	metricapi.Float64Histogram
}

func (h ctxFloat64Histogram) Record(ctx context.Context, value float64, opts ...metricapi.RecordOption) {
	h.Float64Histogram.Record(ctx, value, recordAttrs(ctx, opts)...)
}

type ctxFloat64Gauge struct {
	metricapi.Float64Gauge
}

func (g ctxFloat64Gauge) Record(ctx context.Context, value float64, opts ...metricapi.RecordOption) {
	g.Float64Gauge.Record(ctx, value, recordAttrs(ctx, opts)...)
}
