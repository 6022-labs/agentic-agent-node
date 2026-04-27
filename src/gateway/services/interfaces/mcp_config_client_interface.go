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
	List(ctx context.Context) (*runtime_models.MCPConnectionListPayload, error)
	Create(ctx context.Context, body *runtime_models.CreateMCPConnectionPayload) (*runtime_models.MCPConnectionPayload, error)
	GetById(ctx context.Context, id uuid.UUID) (*runtime_models.MCPConnectionPayload, error)
	Update(ctx context.Context, id uuid.UUID, body *runtime_models.UpdateMCPConnectionPayload) (*runtime_models.MCPConnectionPayload, error)
	Delete(ctx context.Context, id uuid.UUID) error
}
