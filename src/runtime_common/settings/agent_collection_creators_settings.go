package settings

import (
	_ "embed"
	"encoding/json"
	"strconv"
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

// AgentCollectionCreatorsSettings holds the AgentCollectionCreator address per
// chain id. Inner join of the built-in registry and the configured chains: a
// chain present here is mintable.
type AgentCollectionCreatorsSettings map[string]common.Address

func NewAgentCollectionCreatorsSettings(logger *zap.Logger, chainsSettings *ChainsSettings) *AgentCollectionCreatorsSettings {
	var f agentCollectionCreatorsFile
	if err := json.Unmarshal(agentCollectionCreatorsJson, &f); err != nil {
		logger.Fatal("failed to parse built-in agent_collection_creators.json", zap.Error(err))
	}

	settings := AgentCollectionCreatorsSettings{}
	for _, c := range f.Creators {
		addr := strings.TrimSpace(c.Address)
		chainId := strconv.FormatUint(c.ChainId, 10)
		if len(addr) == 0 || !chainsSettings.Has(chainId) {
			continue
		}
		settings[chainId] = common.HexToAddress(addr)
	}
	return &settings
}

// AddressFor returns the AgentCollectionCreator address for a chain, if any.
func (s *AgentCollectionCreatorsSettings) AddressFor(chainId string) (common.Address, bool) {
	if s == nil {
		return common.Address{}, false
	}
	addr, ok := (*s)[chainId]
	return addr, ok
}
