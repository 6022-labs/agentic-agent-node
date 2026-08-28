package metrics

import (
	"context"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/metric"
	"go.opentelemetry.io/otel/metric/noop"
)

const meterName = "agentic_runtime_knowledge"

type KnowledgeMetrics struct {
	conversationStartedCounter metric.Int64Counter
	conversationNamedCounter   metric.Int64Counter
	messageAppendedCounter     metric.Int64Counter
}

func NewKnowledgeMetrics() *KnowledgeMetrics {
	noopMeter := noop.NewMeterProvider().Meter(meterName)
	meter := otel.GetMeterProvider().Meter(meterName)

	conversationStartedCounter, err := meter.Int64Counter(
		"knowledge.conversation.started",
		metric.WithDescription("Number of conversations started"),
	)
	if err != nil {
		conversationStartedCounter, _ = noopMeter.Int64Counter("knowledge.conversation.started")
	}

	conversationNamedCounter, err := meter.Int64Counter(
		"knowledge.conversation.named",
		metric.WithDescription("Number of conversations labelled"),
	)
	if err != nil {
		conversationNamedCounter, _ = noopMeter.Int64Counter("knowledge.conversation.named")
	}

	messageAppendedCounter, err := meter.Int64Counter(
		"knowledge.message.appended",
		metric.WithDescription("Number of conversation messages appended"),
	)
	if err != nil {
		messageAppendedCounter, _ = noopMeter.Int64Counter("knowledge.message.appended")
	}

	return &KnowledgeMetrics{
		conversationStartedCounter: conversationStartedCounter,
		conversationNamedCounter:   conversationNamedCounter,
		messageAppendedCounter:     messageAppendedCounter,
	}
}

func (m *KnowledgeMetrics) RecordConversationStarted(ctx context.Context) {
	m.conversationStartedCounter.Add(ctx, 1)
}

func (m *KnowledgeMetrics) RecordConversationNamed(ctx context.Context) {
	m.conversationNamedCounter.Add(ctx, 1)
}

func (m *KnowledgeMetrics) RecordMessageAppended(ctx context.Context, role string, count int64) {
	m.messageAppendedCounter.Add(ctx, count, metric.WithAttributes(attribute.String("role", role)))
}
