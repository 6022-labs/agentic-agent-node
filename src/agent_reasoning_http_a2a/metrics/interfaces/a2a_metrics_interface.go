package interfaces

import (
	"context"
	"time"
)

// A2AMetricsInterface records outbound agent-to-agent calls; callers pass the raw outcome.
type A2AMetricsInterface interface {
	// RecordSendMessage records one SendMessage turn; paymentRequired marks a 402 answer.
	RecordSendMessage(ctx context.Context, endpointUrl string, duration time.Duration, err error, paymentRequired bool)
	// RecordFetchAgentCard records one agent-card read.
	RecordFetchAgentCard(ctx context.Context, baseUrl string, duration time.Duration, err error)
}
