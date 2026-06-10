package settings

import (
	_ "embed"
	"encoding/json"
	"strings"

	"github.com/ethereum/go-ethereum/common"
	"go.uber.org/zap"
)

//go:embed agent_collection_creators.json
var agentCollectionCreatorsJson []byte

type agentCollectionCreatorEntry struct {
	ChainId uint64 `json:"chainId"`
	Name    string `json:"name"`
	Address string `json:"address"`
}

type agentCollectionCreatorsFile struct {
	Creators []agentCollectionCreatorEntry `json:"creators"`
}

// AgentCollectionCreatorsSettings holds the AgentCollectionCreator address
// for this agent's chain, bound at startup from BlockchainSettings.Id.
type AgentCollectionCreatorsSettings struct {
	CreatorAddress common.Address
}

func NewAgentCollectionCreatorsSettings(
	logger *zap.Logger,
	blockchainSettings *BlockchainSettings,
) *AgentCollectionCreatorsSettings {
	var f agentCollectionCreatorsFile
	if err := json.Unmarshal(agentCollectionCreatorsJson, &f); err != nil {
		logger.Fatal("failed to parse built-in agent_collection_creators.json", zap.Error(err))
	}

	for _, c := range f.Creators {
		if c.ChainId != blockchainSettings.Id {
			continue
		}
		addr := strings.TrimSpace(c.Address)
		if len(addr) == 0 {
			logger.Fatal(
				"AgentCollectionCreator address is empty for the configured chain; populate agent_collection_creators.json and rebuild",
				zap.Uint64("chainId", blockchainSettings.Id),
				zap.String("name", c.Name),
			)
		}
		logger.Info("resolved AgentCollectionCreator from built-in registry",
			zap.Uint64("chainId", blockchainSettings.Id),
			zap.String("name", c.Name),
			zap.String("address", addr),
		)
		return &AgentCollectionCreatorsSettings{
			CreatorAddress: common.HexToAddress(addr),
		}
	}

	logger.Fatal(
		"no AgentCollectionCreator entry for the configured chain; add an entry to agent_collection_creators.json and rebuild",
		zap.Uint64("chainId", blockchainSettings.Id),
	)
	return nil
}
