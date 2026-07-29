package interfaces

import (
	"context"
	"time"
)

// CompletionCall carries the gen_ai.* attributes shared by the client metrics.
type CompletionCall struct {
	Provider      string
	RequestModel  string
	ResponseModel string
}

// LlmInferenceMetricsInterface records OTel GenAI-semconv metrics (gen_ai.*) for LLM chat-completion calls.
type LlmInferenceMetricsInterface interface {
	RecordOperationDuration(ctx context.Context, call CompletionCall, duration time.Duration, errorType string)

	RecordTokenUsage(ctx context.Context, call CompletionCall, tokenType string, tokens int64)

	// RecordProbe records a connection validation, which is a real chat completion underneath.
	RecordProbe(ctx context.Context, call CompletionCall, duration time.Duration, failed bool)
}
