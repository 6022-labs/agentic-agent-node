package interfaces

import (
	"context"

	runtime_models "github.com/6022-labs/agentic-agent-node/src/gateway_http_runtime/models"
)

type ResponseClientInterface interface {
	Create(ctx context.Context, body *runtime_models.ExternalCreateResponsePayload) (*runtime_models.ExternalCreateResponseResultPayload, error)
}
