package observability

import (
	"context"

	sdktrace "go.opentelemetry.io/otel/sdk/trace"
)

// NewAgentIdentitySpanProcessor stamps the ctx agent identity on every span at start.
func NewAgentIdentitySpanProcessor() sdktrace.SpanProcessor {
	return agentIdentitySpanProcessor{}
}

type agentIdentitySpanProcessor struct{}

func (agentIdentitySpanProcessor) OnStart(ctx context.Context, span sdktrace.ReadWriteSpan) {
	if identity, ok := AgentIdentityFromContext(ctx); ok {
		if attrs := identity.attributes(); len(attrs) != 0 {
			span.SetAttributes(attrs...)
		}
	}
}

func (agentIdentitySpanProcessor) OnEnd(sdktrace.ReadOnlySpan)      {}
func (agentIdentitySpanProcessor) Shutdown(context.Context) error   { return nil }
func (agentIdentitySpanProcessor) ForceFlush(context.Context) error { return nil }
