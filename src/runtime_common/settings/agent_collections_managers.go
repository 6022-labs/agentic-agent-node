package settings

import (
	_ "embed"
	"encoding/json"
	"strings"

	"github.com/ethereum/go-ethereum/common"
	"go.uber.org/zap"
)

//go:embed agent_collections_managers.json
var agentCollectionsManagersJson []byte

type agentCollectionsManagerEntry struct {
	ChainId uint64 `json:"chainId"`
	Name    string `json:"name"`
	Address string `json:"address"`
}

type agentCollectionsManagersFile struct {
	Managers []agentCollectionsManagerEntry `json:"managers"`
}

// AgentCollectionsManagersSettings holds the AgentCollectionsManager address
// for this agent's chain, bound at startup from BlockchainSettings.Id.
type AgentCollectionsManagersSettings struct {
	ManagerAddress common.Address
}

func NewAgentCollectionsManagersSettings(
	logger *zap.Logger,
	blockchainSettings *BlockchainSettings,
) *AgentCollectionsManagersSettings {
	var f agentCollectionsManagersFile
	if err := json.Unmarshal(agentCollectionsManagersJson, &f); err != nil {
		logger.Fatal("failed to parse built-in agent_collections_managers.json", zap.Error(err))
	}

	for _, m := range f.Managers {
		if m.ChainId != blockchainSettings.Id {
			continue
		}
		addr := strings.TrimSpace(m.Address)
		if len(addr) == 0 {
			logger.Fatal(
				"AgentCollectionsManager address is empty for the configured chain; populate agent_collections_managers.json and rebuild",
				zap.Uint64("chainId", blockchainSettings.Id),
				zap.String("name", m.Name),
			)
		}
		logger.Info("resolved AgentCollectionsManager from built-in registry",
			zap.Uint64("chainId", blockchainSettings.Id),
			zap.String("name", m.Name),
			zap.String("address", addr),
		)
		return &AgentCollectionsManagersSettings{
			ManagerAddress: common.HexToAddress(addr),
		}
	}

	logger.Fatal(
		"no AgentCollectionsManager entry for the configured chain; add an entry to agent_collections_managers.json and rebuild",
		zap.Uint64("chainId", blockchainSettings.Id),
	)
	return nil
}
