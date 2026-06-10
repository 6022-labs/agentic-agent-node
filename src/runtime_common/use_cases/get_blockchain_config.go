package use_cases

import (
	"context"

	"github.com/6022-labs/agentic-agent-node/src/runtime_common/settings"
	"github.com/6022-labs/agentic-agent-node/src/runtime_common/use_cases/responses"
)

// GetBlockchainConfig assembles the chain id and contract addresses the node
// was bound to at startup, so the owner UI configures itself from the API.
type GetBlockchainConfig struct {
	blockchainSettings               *settings.BlockchainSettings
	agentCollectionsManagersSettings *settings.AgentCollectionsManagersSettings
	agentCollectionCreatorsSettings  *settings.AgentCollectionCreatorsSettings
	token6022Settings                *settings.Token6022Settings
	agentEnsRegistriesSettings       *settings.AgentEnsRegistriesSettings
}

func NewGetBlockchainConfig(
	blockchainSettings *settings.BlockchainSettings,
	agentCollectionsManagersSettings *settings.AgentCollectionsManagersSettings,
	agentCollectionCreatorsSettings *settings.AgentCollectionCreatorsSettings,
	token6022Settings *settings.Token6022Settings,
	agentEnsRegistriesSettings *settings.AgentEnsRegistriesSettings,
) *GetBlockchainConfig {
	return &GetBlockchainConfig{
		blockchainSettings:               blockchainSettings,
		agentCollectionsManagersSettings: agentCollectionsManagersSettings,
		agentCollectionCreatorsSettings:  agentCollectionCreatorsSettings,
		token6022Settings:                token6022Settings,
		agentEnsRegistriesSettings:       agentEnsRegistriesSettings,
	}
}

func (u *GetBlockchainConfig) Execute(_ context.Context) (*responses.BlockchainConfigResponse, error) {
	return &responses.BlockchainConfigResponse{
		ChainId:                   u.blockchainSettings.Id,
		CollectionsManagerAddress: u.agentCollectionsManagersSettings.ManagerAddress.Hex(),
		CollectionCreatorAddress:  u.agentCollectionCreatorsSettings.CreatorAddress.Hex(),
		Token6022Address:          u.token6022Settings.SmartContractAddress.Hex(),
		EnsRegistryAddress:        u.agentEnsRegistriesSettings.RegistryAddress.Hex(),
	}, nil
}
