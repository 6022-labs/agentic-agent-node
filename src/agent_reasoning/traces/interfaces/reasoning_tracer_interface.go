package interfaces

import "context"

// ReasoningTracerInterface opens gen_ai-convention spans for the inference loop.
// Implementations own all OTel details so callers stay vendor-neutral.
type ReasoningTracerInterface interface {
	// StartInference opens the root invoke_agent span for one /responses attempt.
	StartInference(ctx context.Context, conversationId, requestModel string) (context.Context, Span)
	// StartChatCompletion opens a chat span for one LLM round-trip.
	StartChatCompletion(ctx context.Context, provider, model string) (context.Context, ChatCompletionSpan)
	// StartToolCall opens an execute_tool span; serverName is empty for provider-run tools.
	StartToolCall(ctx context.Context, toolName, toolCallId, serverName string) (context.Context, Span)
	// StartListTools opens an MCP tools/list span for tool discovery.
	StartListTools(ctx context.Context, serverName string) (context.Context, Span)
}
