package settings

import (
	_ "embed"
	"encoding/json"
	"slices"
	"strconv"
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

// AgentCollectionsManagersSettings holds the AgentCollectionsManager address per
// chain id (inner join of the built-in registry and the configured chains).
type AgentCollectionsManagersSettings map[string]common.Address

func NewAgentCollectionsManagersSettings(
	logger *zap.Logger,
	evmChainsSettings *EvmChainsSettings,
) *AgentCollectionsManagersSettings {
	var f agentCollectionsManagersFile
	if err := json.Unmarshal(agentCollectionsManagersJson, &f); err != nil {
		logger.Fatal("failed to parse built-in agent_collections_managers.json", zap.Error(err))
	}

	settings := AgentCollectionsManagersSettings{}
	for _, m := range f.Managers {
		addr := strings.TrimSpace(m.Address)
		chainId := strconv.FormatUint(m.ChainId, 10)
		if len(addr) == 0 || !evmChainsSettings.Has(chainId) {
			continue
		}
		settings[chainId] = common.HexToAddress(addr)
	}
	return &settings
}

// AddressFor returns the AgentCollectionsManager address for a chain, if any.
func (s *AgentCollectionsManagersSettings) AddressFor(chainId string) (common.Address, bool) {
	if s == nil {
		return common.Address{}, false
	}
	addr, ok := (*s)[chainId]
	return addr, ok
}

// ChainIds lists the chains with a configured manager, sorted (map iteration is unordered).
func (s *AgentCollectionsManagersSettings) ChainIds() []string {
	if s == nil {
		return nil
	}
	chainIds := make([]string, 0, len(*s))
	for chainId := range *s {
		chainIds = append(chainIds, chainId)
	}
	slices.Sort(chainIds)
	return chainIds
}
