package traces

import (
	"context"

	"github.com/6022-labs/agentic-agent-node/src/agent_reasoning/traces/interfaces"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
)

const (
	tracerName = "agentic_runtime_reasoning"
	// toolTypeFunction is the gen_ai.tool.type value for callable tools (vs extension/datastore).
	toolTypeFunction = "function"
)

// conversationIdKey carries the conversation id down ctx so child spans can tag
// gen_ai.conversation.id without threading it through every call.
type conversationIdKey struct{}

func NewReasoningTracer() *ReasoningTracer {
	return &ReasoningTracer{tracer: otel.Tracer(tracerName)}
}

type ReasoningTracer struct {
	tracer trace.Tracer
}

func (t *ReasoningTracer) conversationIdFromContext(ctx context.Context) string {
	id, _ := ctx.Value(conversationIdKey{}).(string)
	return id
}

func (t *ReasoningTracer) StartInference(ctx context.Context, conversationId, requestModel string) (context.Context, interfaces.Span) {
	ctx = context.WithValue(ctx, conversationIdKey{}, conversationId)
	ctx, span := t.tracer.Start(ctx, "invoke_agent")
	span.SetAttributes(
		attribute.String("gen_ai.operation.name", "invoke_agent"),
		attribute.String("gen_ai.conversation.id", conversationId),
		// langfuse.session.id groups a conversation's traces in Langfuse; inert elsewhere.
		attribute.String("langfuse.session.id", conversationId),
	)
	if requestModel != "" {
		span.SetAttributes(attribute.String("gen_ai.request.model", requestModel))
	}
	return ctx, &otelSpan{span: span}
}

func (t *ReasoningTracer) StartChatCompletion(ctx context.Context, provider, model string) (context.Context, interfaces.ChatCompletionSpan) {
	ctx, span := t.tracer.Start(ctx, "chat "+model, trace.WithSpanKind(trace.SpanKindClient))
	span.SetAttributes(
		attribute.String("gen_ai.operation.name", "chat"),
		attribute.String("gen_ai.provider.name", provider),
		attribute.String("gen_ai.request.model", model),
	)
	if conversationId := t.conversationIdFromContext(ctx); conversationId != "" {
		span.SetAttributes(attribute.String("gen_ai.conversation.id", conversationId))
	}
	return ctx, &chatCompletionSpan{otelSpan{span: span}}
}

func (t *ReasoningTracer) StartToolCall(ctx context.Context, toolName, toolCallId, serverName string) (context.Context, interfaces.Span) {
	ctx, span := t.tracer.Start(ctx, "execute_tool "+toolName)
	span.SetAttributes(
		attribute.String("gen_ai.operation.name", "execute_tool"),
		attribute.String("gen_ai.tool.name", toolName),
		attribute.String("gen_ai.tool.type", toolTypeFunction),
	)
	if toolCallId != "" {
		span.SetAttributes(attribute.String("gen_ai.tool.call.id", toolCallId))
	}
	if serverName != "" {
		span.SetAttributes(attribute.String("mcp.server.name", serverName))
	}
	if conversationId := t.conversationIdFromContext(ctx); conversationId != "" {
		span.SetAttributes(attribute.String("gen_ai.conversation.id", conversationId))
	}
	return ctx, &otelSpan{span: span}
}

func (t *ReasoningTracer) StartListTools(ctx context.Context, serverName string) (context.Context, interfaces.Span) {
	ctx, span := t.tracer.Start(ctx, "tools/list", trace.WithSpanKind(trace.SpanKindClient))
	span.SetAttributes(
		attribute.String("mcp.method.name", "tools/list"),
		attribute.String("mcp.server.name", serverName),
	)
	return ctx, &otelSpan{span: span}
}
