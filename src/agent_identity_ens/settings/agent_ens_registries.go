package settings

import (
	_ "embed"
	"encoding/json"
	"strings"

	blockchain_settings "github.com/6022-labs/agentic-agent-node/src/agent_identity_blockchain/settings"
	"github.com/ethereum/go-ethereum/common"
	"go.uber.org/zap"
)

//go:embed agent_ens_registries.json
var agentEnsRegistriesJson []byte

type agentEnsRegistryEntry struct {
	ChainId uint64 `json:"chainId"`
	Name    string `json:"name"`
	Address string `json:"address"`
}

type agentEnsRegistriesFile struct {
	Registries []agentEnsRegistryEntry `json:"registries"`
}

// AgentEnsRegistriesSettings holds the AgentEnsRegistry contract address bound
// to the chain this agent runs against. Resolved at startup from the built-in
// registry; fail-fast if no entry exists for the configured chain.
type AgentEnsRegistriesSettings struct {
	RegistryAddress common.Address
}

func NewAgentEnsRegistriesSettings(
	logger *zap.Logger,
	blockchainSettings *blockchain_settings.BlockchainSettings,
) *AgentEnsRegistriesSettings {
	var f agentEnsRegistriesFile
	if err := json.Unmarshal(agentEnsRegistriesJson, &f); err != nil {
		logger.Fatal("failed to parse built-in agent_ens_registries.json", zap.Error(err))
	}

	for _, w := range f.Registries {
		if w.ChainId != blockchainSettings.Id {
			continue
		}
		addr := strings.TrimSpace(w.Address)
		if len(addr) == 0 {
			logger.Fatal(
				"AgentEnsRegistry address is empty for the configured chain; populate agent_ens_registries.json and rebuild",
				zap.Uint64("chainId", blockchainSettings.Id),
				zap.String("name", w.Name),
			)
		}
		logger.Info("resolved AgentEnsRegistry from built-in registry",
			zap.Uint64("chainId", blockchainSettings.Id),
			zap.String("name", w.Name),
			zap.String("address", addr),
		)
		return &AgentEnsRegistriesSettings{
			RegistryAddress: common.HexToAddress(addr),
		}
	}

	logger.Fatal(
		"no AgentEnsRegistry entry for the configured chain; add an entry to agent_ens_registries.json and rebuild",
		zap.Uint64("chainId", blockchainSettings.Id),
	)
	return nil
}
