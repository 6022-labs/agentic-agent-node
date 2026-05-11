package interfaces

import (
	"context"
	"errors"

	runtime_models "github.com/6022-labs/agentic-agent-node/src/gateway_http_runtime/models"
)

// ErrLLMConnectionNotFound is returned by Get when no LLM connection exists.
// Part of the contract — every impl must use this exact sentinel.
var ErrLLMConnectionNotFound = errors.New("llm connection not configured")

// LLMConfigClientInterface is the gateway-side port for the runtime's LLM
// config endpoints. Implemented by `gateway_http_runtime/services` over HTTP;
// future in-process adapters (e.g. MCP-driven self-config) plug in here too.
type LLMConfigClientInterface interface {
	Get(ctx context.Context) (*runtime_models.ExternalLLMConnectionPayload, error)
	Setup(ctx context.Context, body *runtime_models.ExternalSetupLLMConnectionPayload) (*runtime_models.ExternalLLMConnectionPayload, error)
	Delete(ctx context.Context) error
}
