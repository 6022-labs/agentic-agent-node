package interfaces

import (
	"context"
)

// CommandMetricsInterface captures what the agent runs on its machine: how commands end, and detached-process churn.
type CommandMetricsInterface interface {
	RecordBackgroundStarted(ctx context.Context)
	RecordProcessKilled(ctx context.Context)
}
