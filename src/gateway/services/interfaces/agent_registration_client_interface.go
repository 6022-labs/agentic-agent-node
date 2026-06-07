package interfaces

import (
	"context"

	runtime_models "github.com/6022-labs/agentic-agent-node/src/gateway_http_runtime/models"
)

// AgentRegistrationClientInterface is the gateway-side port for the runtime's
// agent mint + ENS registration endpoints. Implemented over HTTP.
type AgentRegistrationClientInterface interface {
	Save(ctx context.Context, body *runtime_models.ExternalSaveAgentRegistrationPayload) (*runtime_models.ExternalAgentRegistrationPayload, error)
	Get(ctx context.Context) (*runtime_models.ExternalAgentRegistrationPayload, error)
	MintInfo(ctx context.Context) (*runtime_models.ExternalMintInfoPayload, error)
	CheckMint(ctx context.Context) (*runtime_models.ExternalRegistrationStatusPayload, error)
	EnsInfo(ctx context.Context) (*runtime_models.ExternalEnsInfoPayload, error)
	CheckEns(ctx context.Context) (*runtime_models.ExternalRegistrationStatusPayload, error)
	Self(ctx context.Context) (*runtime_models.ExternalRegistrationStatusPayload, error)
}
