package interfaces

import "context"

// ResponseMetricsInterface records agent-response generation outcomes.
type ResponseMetricsInterface interface {
	RecordResponseSucceeded(ctx context.Context)
	RecordResponseFailed(ctx context.Context)
}
