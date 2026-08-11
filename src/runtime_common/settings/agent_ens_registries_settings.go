package settings

import (
	_ "embed"
	"encoding/json"
	"strconv"
	"strings"

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

// AgentEnsRegistriesSettings holds the AgentEnsRegistry contract address per
// chain id (inner join of the built-in registry and the configured chains).
type AgentEnsRegistriesSettings map[string]common.Address

func NewAgentEnsRegistriesSettings(
	logger *zap.Logger,
	evmChainsSettings *EvmChainsSettings,
) *AgentEnsRegistriesSettings {
	var f agentEnsRegistriesFile
	if err := json.Unmarshal(agentEnsRegistriesJson, &f); err != nil {
		logger.Fatal("failed to parse built-in agent_ens_registries.json", zap.Error(err))
	}

	settings := AgentEnsRegistriesSettings{}
	for _, w := range f.Registries {
		addr := strings.TrimSpace(w.Address)
		chainId := strconv.FormatUint(w.ChainId, 10)
		if len(addr) == 0 || !evmChainsSettings.Has(chainId) {
			continue
		}
		settings[chainId] = common.HexToAddress(addr)
	}
	return &settings
}

// AddressFor returns the AgentEnsRegistry address for a chain, if any.
func (s *AgentEnsRegistriesSettings) AddressFor(chainId string) (common.Address, bool) {
	if s == nil {
		return common.Address{}, false
	}
	addr, ok := (*s)[chainId]
	return addr, ok
}
