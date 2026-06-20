package interfaces

import "context"

// AuthorizationMetricsInterface records agent authorization-check outcomes.
type AuthorizationMetricsInterface interface {
	RecordAuthorizationAuthorized(ctx context.Context)
	RecordAuthorizationDenied(ctx context.Context)
	RecordAuthorizationFailed(ctx context.Context)
}
