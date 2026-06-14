package interfaces

import (
	"context"

	"github.com/6022-labs/agentic-agent-node/src/gateway/identity/models"
)

// AgentRegistrationClientInterface is the gateway-side port for the runtime's
// agent mint + ENS registration endpoints. Implemented over HTTP.
type AgentRegistrationClientInterface interface {
	Save(ctx context.Context, body *models.ExternalSaveAgentRegistrationPayload) (*models.ExternalAgentRegistrationPayload, error)
	Get(ctx context.Context) (*models.ExternalAgentRegistrationPayload, error)
	MintInfo(ctx context.Context) (*models.ExternalMintInfoPayload, error)
	CheckMint(ctx context.Context) (*models.ExternalRegistrationStatusPayload, error)
	EnsInfo(ctx context.Context) (*models.ExternalEnsInfoPayload, error)
	CheckEns(ctx context.Context) (*models.ExternalRegistrationStatusPayload, error)
	Self(ctx context.Context) (*models.ExternalRegistrationStatusPayload, error)
}
