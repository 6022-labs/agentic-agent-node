package interfaces

import (
	"context"
	"errors"

	runtime_models "github.com/6022-labs/agentic-agent-node/src/gateway_http_runtime/models"
	"github.com/google/uuid"
)

// ErrMCPConnectionNotFound is returned by GetById/Update when no row matches.
var ErrMCPConnectionNotFound = errors.New("mcp connection not found")

// MCPConfigClientInterface is the gateway-side port for the runtime's MCP
// config endpoints. Implemented by `gateway_http_runtime/services` over HTTP.
type MCPConfigClientInterface interface {
	List(ctx context.Context) (*runtime_models.ExternalMCPConnectionListPayload, error)
	Create(ctx context.Context, body *runtime_models.ExternalCreateMCPConnectionPayload) (*runtime_models.ExternalMCPConnectionPayload, error)
	GetById(ctx context.Context, id uuid.UUID) (*runtime_models.ExternalMCPConnectionPayload, error)
	Update(ctx context.Context, id uuid.UUID, body *runtime_models.ExternalUpdateMCPConnectionPayload) (*runtime_models.ExternalMCPConnectionPayload, error)
	Delete(ctx context.Context, id uuid.UUID) error
}
