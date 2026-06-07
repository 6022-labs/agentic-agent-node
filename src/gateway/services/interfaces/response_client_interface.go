package interfaces

import (
	"context"

	"github.com/6022-labs/agentic-agent-node/src/gateway/models"
)

type ResponseClientInterface interface {
	Create(ctx context.Context, body *models.ExternalCreateResponsePayload) (*models.ExternalCreateResponseResultPayload, error)
}
