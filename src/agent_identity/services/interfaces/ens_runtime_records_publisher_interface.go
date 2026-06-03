package interfaces

import "context"

// EnsRuntimeRecordsPublisherInterface writes the agent's runtime ENS text
// records (URL, kind). Failure is non-fatal — the caller logs and continues.
type EnsRuntimeRecordsPublisherInterface interface {
	PublishRuntimeRecords(ctx context.Context) error
}
