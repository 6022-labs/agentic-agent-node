package use_cases

import (
	"context"

	"github.com/6022-labs/agentic-agent-node/src/gateway/core/models"
	gateway_interfaces "github.com/6022-labs/agentic-agent-node/src/gateway/core/services/interfaces"
	"github.com/6022-labs/agentic-agent-node/src/gateway/core/use_cases/responses"
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
	chains := make(map[string]responses.ChainConfig, len(p.Chains))
	for chainId, c := range p.Chains {
		chains[chainId] = responses.ChainConfig{
			CollectionsManagerAddress: c.CollectionsManagerAddress,
			CollectionCreatorAddress:  c.CollectionCreatorAddress,
			Token6022Address:          c.Token6022Address,
			EnsRegistryAddress:        c.EnsRegistryAddress,
			Mintable:                  c.Mintable,
			FacilitatorNetwork:        c.FacilitatorNetwork,
		}
	}
	return &responses.BlockchainConfigResponse{Chains: chains}
}
