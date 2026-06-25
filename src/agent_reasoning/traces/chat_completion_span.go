package traces

import (
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
)

// chatCompletionSpan is an otelSpan that also records token usage.
type chatCompletionSpan struct {
	otelSpan
}

func (s *chatCompletionSpan) RecordUsage(inputTokens, outputTokens int) {
	s.span.SetAttributes(
		attribute.Int("gen_ai.usage.input_tokens", inputTokens),
		attribute.Int("gen_ai.usage.output_tokens", outputTokens),
	)
}

func (s *chatCompletionSpan) RecordToolCall(toolName, toolCallId string) {
	s.span.AddEvent("execute_tool", trace.WithAttributes(
		attribute.String("gen_ai.operation.name", "execute_tool"),
		attribute.String("gen_ai.tool.name", toolName),
		attribute.String("gen_ai.tool.type", toolTypeFunction),
		attribute.String("gen_ai.tool.call.id", toolCallId),
	))
}

func (s *chatCompletionSpan) RecordResponse(model string, finishReasons []string, responseId string) {
	if model != "" {
		s.span.SetAttributes(attribute.String("gen_ai.response.model", model))
	}
	if len(finishReasons) > 0 {
		s.span.SetAttributes(attribute.StringSlice("gen_ai.response.finish_reasons", finishReasons))
	}
	if responseId != "" {
		s.span.SetAttributes(attribute.String("gen_ai.response.id", responseId))
	}
}
