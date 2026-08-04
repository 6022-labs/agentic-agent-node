package interfaces

import (
	"context"

	"github.com/google/uuid"
)

// KillProcessExecutorInterface terminates a detached command's process group.
type KillProcessExecutorInterface interface {
	Kill(ctx context.Context, id uuid.UUID) error
}
