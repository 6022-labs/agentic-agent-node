package interfaces

import "context"

// RegistrationMetricsInterface records agent self-registration outcomes.
type RegistrationMetricsInterface interface {
	RecordRegistrationSucceeded(ctx context.Context)
	RecordRegistrationPending(ctx context.Context)
	RecordRegistrationFailed(ctx context.Context)
}
