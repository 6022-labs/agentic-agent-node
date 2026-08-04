package metrics

import (
	"context"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/metric"
	"go.opentelemetry.io/otel/metric/noop"
)

const commandMeterName = "agentic_runtime_reasoning_os"

// Command duration is not recorded here: the self tool registry already times every call as gen_ai.execute_tool.duration.
type CommandMetrics struct {
	backgroundStartedCounter metric.Int64Counter
	processKilledCounter     metric.Int64Counter
}

func NewCommandMetrics() *CommandMetrics {
	noopMeter := noop.NewMeterProvider().Meter(commandMeterName)
	meter := otel.GetMeterProvider().Meter(commandMeterName)

	backgroundStartedCounter, err := meter.Int64Counter(
		"reasoning.command.background_started",
		metric.WithUnit("{command}"),
		metric.WithDescription("Number of commands the agent started in the background"),
	)
	if err != nil {
		backgroundStartedCounter, _ = noopMeter.Int64Counter("reasoning.command.background_started")
	}

	processKilledCounter, err := meter.Int64Counter(
		"reasoning.command.killed",
		metric.WithUnit("{process}"),
		metric.WithDescription("Number of background processes killed on request"),
	)
	if err != nil {
		processKilledCounter, _ = noopMeter.Int64Counter("reasoning.command.killed")
	}

	return &CommandMetrics{
		backgroundStartedCounter: backgroundStartedCounter,
		processKilledCounter:     processKilledCounter,
	}
}

func (m *CommandMetrics) RecordBackgroundStarted(ctx context.Context) {
	m.backgroundStartedCounter.Add(ctx, 1)
}

func (m *CommandMetrics) RecordProcessKilled(ctx context.Context) {
	m.processKilledCounter.Add(ctx, 1)
}
