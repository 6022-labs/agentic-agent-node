package use_cases

import (
	"context"

	"github.com/6022-labs/agentic-agent-node/src/gateway/core/models"
	gateway_interfaces "github.com/6022-labs/agentic-agent-node/src/gateway/core/services/interfaces"
	"github.com/6022-labs/agentic-agent-node/src/gateway/core/settings"
	"github.com/6022-labs/agentic-agent-node/src/gateway/core/use_cases/responses"
)

// GetConfig assembles the public node config: the runtime's chain config plus the
// IPFS gateways the console resolves CIDs with.
type GetConfig struct {
	blockchainConfigClient gateway_interfaces.BlockchainConfigClientInterface
	ipfsGatewaysSettings   *settings.IPFSGatewaysSettings
}

func NewGetConfig(
	blockchainConfigClient gateway_interfaces.BlockchainConfigClientInterface,
	ipfsGatewaysSettings *settings.IPFSGatewaysSettings,
) *GetConfig {
	return &GetConfig{
		blockchainConfigClient: blockchainConfigClient,
		ipfsGatewaysSettings:   ipfsGatewaysSettings,
	}
}

func (u *GetConfig) Execute(ctx context.Context) (*responses.ConfigResponse, error) {
	payload, err := u.blockchainConfigClient.Get(ctx)
	if err != nil {
		return nil, err
	}
	return u.toResponse(payload), nil
}

func (u *GetConfig) toResponse(
	p *models.ExternalBlockchainConfigPayload,
) *responses.ConfigResponse {
	evmChains := make(map[string]responses.EvmChainConfig, len(p.EvmChains))
	for chainId, c := range p.EvmChains {
		evmChains[chainId] = responses.EvmChainConfig{
			CollectionsManagerAddress: c.CollectionsManagerAddress,
			CollectionCreatorAddress:  c.CollectionCreatorAddress,
			Token6022Address:          c.Token6022Address,
			EnsRegistryAddress:        c.EnsRegistryAddress,
			Mintable:                  c.Mintable,
			FacilitatorNetwork:        c.FacilitatorNetwork,
		}
	}
	return &responses.ConfigResponse{
		EvmChains:               evmChains,
		IpfsGatewayUrlTemplates: u.ipfsGatewaysSettings.GatewayUrlTemplates,
	}
}
