package interfaces

import "context"

// BootstrapMetricsInterface records agent bootstrap outcomes.
type BootstrapMetricsInterface interface {
	RecordBootstrapSucceeded(ctx context.Context)
	RecordBootstrapFailed(ctx context.Context)
}
