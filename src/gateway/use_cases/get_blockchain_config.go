package use_cases

import (
	"context"
	"sync"

	"github.com/6022-labs/agentic-agent-node/src/gateway/models"
	gateway_interfaces "github.com/6022-labs/agentic-agent-node/src/gateway/services/interfaces"
	"github.com/6022-labs/agentic-agent-node/src/gateway/use_cases/responses"
)

// GetBlockchainConfig proxies the runtime's static chain config, caching the
// first successful fetch for the life of the process (errors are never cached).
type GetBlockchainConfig struct {
	blockchainConfigClient gateway_interfaces.BlockchainConfigClientInterface

	mu     sync.Mutex
	cached *responses.BlockchainConfigResponse
}

func NewGetBlockchainConfig(
	blockchainConfigClient gateway_interfaces.BlockchainConfigClientInterface,
) *GetBlockchainConfig {
	return &GetBlockchainConfig{blockchainConfigClient: blockchainConfigClient}
}

func (u *GetBlockchainConfig) Execute(ctx context.Context) (*responses.BlockchainConfigResponse, error) {
	u.mu.Lock()
	defer u.mu.Unlock()

	if u.cached != nil {
		return u.cached, nil
	}

	payload, err := u.blockchainConfigClient.Get(ctx)
	if err != nil {
		return nil, err
	}

	u.cached = u.toResponse(payload)
	return u.cached, nil
}

func (u *GetBlockchainConfig) toResponse(p *models.ExternalBlockchainConfigPayload) *responses.BlockchainConfigResponse {
	return &responses.BlockchainConfigResponse{
		ChainId:                   p.ChainId,
		CollectionsManagerAddress: p.CollectionsManagerAddress,
		CollectionCreatorAddress:  p.CollectionCreatorAddress,
		Token6022Address:          p.Token6022Address,
		EnsRegistryAddress:        p.EnsRegistryAddress,
	}
}
