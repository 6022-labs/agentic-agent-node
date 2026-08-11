package interfaces

import (
	"context"
	"time"
)

// SelfToolMetricsInterface records the OTel GenAI-semconv tool metric; callers pass the raw error, not semconv values.
type SelfToolMetricsInterface interface {
	// RecordToolCall records one self tool call; err is nil on success.
	RecordToolCall(ctx context.Context, toolName string, duration time.Duration, err error)
}
