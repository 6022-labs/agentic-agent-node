package interfaces

import (
	"context"
	"time"
)

// ReasoningMetricsInterface captures inference-loop health: run duration, iterations, tool calls, and ceiling hits.
type ReasoningMetricsInterface interface {
	RecordLoopCompleted(ctx context.Context, duration time.Duration, iterations int64)
	RecordLoopExceeded(ctx context.Context, duration time.Duration, iterations int64)
	// RecordLoopAborted records a loop killed by a non-iteration safety ceiling
	// (reason: deadline_exceeded, token_budget_exceeded).
	RecordLoopAborted(ctx context.Context, reason string, duration time.Duration, iterations int64)
	RecordToolCalls(ctx context.Context, count int64)
}
