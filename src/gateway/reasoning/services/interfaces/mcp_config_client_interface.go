package interfaces

import (
	"context"
	"errors"

	"github.com/6022-labs/agentic-agent-node/src/gateway/reasoning/models"
	"github.com/google/uuid"
)

// ErrMCPConnectionNotFound is returned by GetById/Update when no row matches.
var ErrMCPConnectionNotFound = errors.New("mcp connection not found")

// MCPConfigClientInterface is the gateway-side port for the runtime's MCP
// config endpoints. Implemented by `gateway_http_runtime/services` over HTTP.
type MCPConfigClientInterface interface {
	List(ctx context.Context) (*models.ExternalMCPConnectionListPayload, error)
	Create(ctx context.Context, body *models.ExternalCreateMCPConnectionPayload) (*models.ExternalMCPConnectionPayload, error)
	GetById(ctx context.Context, id uuid.UUID) (*models.ExternalMCPConnectionPayload, error)
	Update(ctx context.Context, id uuid.UUID, body *models.ExternalUpdateMCPConnectionPayload) (*models.ExternalMCPConnectionPayload, error)
	Delete(ctx context.Context, id uuid.UUID) error
}
