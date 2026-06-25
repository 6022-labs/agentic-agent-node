package interfaces

// ChatCompletionSpan is a Span carrying token usage and provider-run tool calls.
type ChatCompletionSpan interface {
	Span
	RecordUsage(inputTokens, outputTokens int)
	// RecordToolCall notes a provider-executed tool as an event (no latency available).
	RecordToolCall(toolName, toolCallId string)
	// RecordResponse tags what the provider returned; empty fields are skipped.
	RecordResponse(model string, finishReasons []string, responseId string)
}
