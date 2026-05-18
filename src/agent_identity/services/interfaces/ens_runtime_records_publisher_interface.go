package interfaces

import "context"

// EnsRuntimeRecordsPublisherInterface writes the agent's runtime ENS text
// records (URL, kind) for its on-chain identity. Invoked by BootstrapRunner
// after the wallet is ready. Failure is operational and non-fatal — the
// caller logs and continues.
type EnsRuntimeRecordsPublisherInterface interface {
	PublishRuntimeRecords(ctx context.Context) error
}
