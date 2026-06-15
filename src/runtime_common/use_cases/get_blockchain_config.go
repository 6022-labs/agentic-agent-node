package use_cases

import (
	"context"
	"strconv"

	"github.com/6022-labs/agentic-agent-node/src/common/helpers"
	"github.com/6022-labs/agentic-agent-node/src/runtime_common/settings"
	"github.com/6022-labs/agentic-agent-node/src/runtime_common/use_cases/responses"
	"github.com/ethereum/go-ethereum/common"
)

// GetBlockchainConfig assembles, per reachable chain, the contract addresses and
// facilitator network the owner UI configures itself from.
type GetBlockchainConfig struct {
	chainsSettings                   *settings.ChainsSettings
	agentCollectionsManagersSettings *settings.AgentCollectionsManagersSettings
	agentCollectionCreatorsSettings  *settings.AgentCollectionCreatorsSettings
	token6022Settings                *settings.Token6022Settings
	agentEnsRegistriesSettings       *settings.AgentEnsRegistriesSettings
	paymentFacilitatorsSettings      *settings.PaymentFacilitatorsSettings
}

func NewGetBlockchainConfig(
	chainsSettings *settings.ChainsSettings,
	agentCollectionsManagersSettings *settings.AgentCollectionsManagersSettings,
	agentCollectionCreatorsSettings *settings.AgentCollectionCreatorsSettings,
	token6022Settings *settings.Token6022Settings,
	agentEnsRegistriesSettings *settings.AgentEnsRegistriesSettings,
	paymentFacilitatorsSettings *settings.PaymentFacilitatorsSettings,
) *GetBlockchainConfig {
	return &GetBlockchainConfig{
		chainsSettings:                   chainsSettings,
		agentCollectionsManagersSettings: agentCollectionsManagersSettings,
		agentCollectionCreatorsSettings:  agentCollectionCreatorsSettings,
		token6022Settings:                token6022Settings,
		agentEnsRegistriesSettings:       agentEnsRegistriesSettings,
		paymentFacilitatorsSettings:      paymentFacilitatorsSettings,
	}
}

func (u *GetBlockchainConfig) Execute(_ context.Context) (*responses.BlockchainConfigResponse, error) {
	chainIds := u.chainsSettings.ChainIds()
	chains := make(map[string]responses.ChainConfig, len(chainIds))
	for _, chainId := range chainIds {
		creator, hasCreator := u.agentCollectionCreatorsSettings.AddressFor(chainId)
		chains[chainId] = responses.ChainConfig{
			CollectionsManagerAddress: hexOrEmpty(u.agentCollectionsManagersSettings.AddressFor(chainId)),
			CollectionCreatorAddress:  hexOrEmpty(creator, hasCreator),
			Token6022Address:          hexOrEmpty(u.token6022Settings.AddressFor(chainId)),
			EnsRegistryAddress:        hexOrEmpty(u.agentEnsRegistriesSettings.AddressFor(chainId)),
			Mintable:                  hasCreator,
			FacilitatorNetwork:        u.facilitatorNetwork(chainId),
		}
	}

	return &responses.BlockchainConfigResponse{Chains: chains}, nil
}

// facilitatorNetwork returns the chain's CAIP-2 network if a facilitator is
// configured for it, else nil.
func (u *GetBlockchainConfig) facilitatorNetwork(chainId string) *string {
	evmChainId, err := strconv.ParseUint(chainId, 10, 64)
	if err != nil {
		return nil
	}
	network := helpers.FormatCaip2(evmChainId)
	if _, ok := u.paymentFacilitatorsSettings.BaseUrlFor(network); !ok {
		return nil
	}
	return &network
}

// hexOrEmpty renders an address as hex, or "" when it isn't configured.
func hexOrEmpty(address common.Address, ok bool) string {
	if !ok {
		return ""
	}
	return address.Hex()
}
