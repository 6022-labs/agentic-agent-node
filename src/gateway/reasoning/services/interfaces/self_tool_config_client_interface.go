package interfaces

import (
	"context"
	"errors"

	"github.com/6022-labs/agentic-agent-node/src/gateway/reasoning/models"
	"github.com/google/uuid"
)

// ErrSelfToolNotFound is returned by Update when no row matches.
var ErrSelfToolNotFound = errors.New("self tool not found")

// SelfToolConfigClientInterface is the gateway-side port for the runtime's
// self tool config endpoints. Implemented by `gateway_http_runtime/services` over HTTP.
type SelfToolConfigClientInterface interface {
	List(ctx context.Context) (*models.ExternalSelfToolListPayload, error)
	Update(ctx context.Context, id uuid.UUID, body *models.ExternalUpdateSelfToolPayload) (*models.ExternalSelfToolPayload, error)
}
