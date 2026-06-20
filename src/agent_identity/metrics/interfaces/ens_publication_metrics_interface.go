package interfaces

import "context"

// EnsPublicationMetricsInterface records ENS runtime-record publication outcomes.
type EnsPublicationMetricsInterface interface {
	RecordEnsPublishedSucceeded(ctx context.Context)
	RecordEnsPublishedFailed(ctx context.Context)
	RecordEnsPublishedNotMinted(ctx context.Context)
}
