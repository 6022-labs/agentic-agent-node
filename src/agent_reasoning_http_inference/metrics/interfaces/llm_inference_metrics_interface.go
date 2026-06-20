package interfaces

import (
	"context"
	"time"
)

// LlmInferenceMetricsInterface records OTel GenAI-semconv metrics (gen_ai.*) for LLM chat-completion calls.
type LlmInferenceMetricsInterface interface {
	RecordOperationDuration(ctx context.Context, provider, model string, duration time.Duration, errorType string)
	RecordTokenUsage(ctx context.Context, provider, model, tokenType string, tokens int64)
}
