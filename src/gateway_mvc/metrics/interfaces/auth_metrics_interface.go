package interfaces

import "context"

type AuthMetricsInterface interface {
	RecordAuthSuccess(ctx context.Context)
	RecordAuthFailure(ctx context.Context)
}
