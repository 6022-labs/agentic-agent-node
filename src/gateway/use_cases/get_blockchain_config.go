package use_cases

import (
	"context"

	"github.com/6022-labs/agentic-agent-node/src/gateway/models"
	gateway_interfaces "github.com/6022-labs/agentic-agent-node/src/gateway/services/interfaces"
	"github.com/6022-labs/agentic-agent-node/src/gateway/use_cases/responses"
)

// GetBlockchainConfig proxies the runtime's chain config.
type GetBlockchainConfig struct {
	blockchainConfigClient gateway_interfaces.BlockchainConfigClientInterface
}

func NewGetBlockchainConfig(
	blockchainConfigClient gateway_interfaces.BlockchainConfigClientInterface,
) *GetBlockchainConfig {
	return &GetBlockchainConfig{blockchainConfigClient: blockchainConfigClient}
}

func (u *GetBlockchainConfig) Execute(ctx context.Context) (*responses.BlockchainConfigResponse, error) {
	payload, err := u.blockchainConfigClient.Get(ctx)
	if err != nil {
		return nil, err
	}
	return u.toResponse(payload), nil
}

func (u *GetBlockchainConfig) toResponse(p *models.ExternalBlockchainConfigPayload) *responses.BlockchainConfigResponse {
	return &responses.BlockchainConfigResponse{
		ChainId:                   p.ChainId,
		CollectionsManagerAddress: p.CollectionsManagerAddress,
		CollectionCreatorAddress:  p.CollectionCreatorAddress,
		Token6022Address:          p.Token6022Address,
		EnsRegistryAddress:        p.EnsRegistryAddress,
		FacilitatorNetworks:       p.FacilitatorNetworks,
	}
}
