package interfaces

import (
	"context"
	"github.com/google/uuid"
)

// DeleteMCPConnectionExecutorInterface exposes the MCP connection delete to the self tools.
type DeleteMCPConnectionExecutorInterface interface {
	Execute(ctx context.Context, id uuid.UUID) error
}
