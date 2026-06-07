package interfaces

import (
	"context"
	"errors"

	"github.com/6022-labs/agentic-agent-node/src/gateway/models"
)

// ErrLLMConnectionNotFound is returned by Get when no LLM connection exists.
// Part of the contract — every impl must use this exact sentinel.
var ErrLLMConnectionNotFound = errors.New("llm connection not configured")

// LLMConfigClientInterface is the gateway-side port for the runtime's LLM
// config endpoints. Implemented over HTTP by gateway_http_runtime/services.
type LLMConfigClientInterface interface {
	Get(ctx context.Context) (*models.ExternalLLMConnectionPayload, error)
	Setup(ctx context.Context, body *models.ExternalSetupLLMConnectionPayload) (*models.ExternalLLMConnectionPayload, error)
	Delete(ctx context.Context) error
}
