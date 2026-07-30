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
	ServerUrl     string
}

// LlmInferenceMetricsInterface records gen_ai.* metrics; callers hand over the raw outcome, not semconv values.
type LlmInferenceMetricsInterface interface {
	// RecordOperationDuration records one chat completion; err is nil on success.
	RecordOperationDuration(ctx context.Context, call CompletionCall, duration time.Duration, err error)

	// RecordTokenUsage records both counts, each on its own gen_ai.token.type series.
	RecordTokenUsage(ctx context.Context, call CompletionCall, inputTokens, outputTokens int64)

	// RecordProbe records a connection validation, which is a real chat completion underneath.
	RecordProbe(ctx context.Context, call CompletionCall, duration time.Duration, err error)
}
