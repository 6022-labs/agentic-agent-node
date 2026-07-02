package settings

import (
	"slices"
	"strings"

	"github.com/knadh/koanf/v2"
	"go.uber.org/zap"
)

const EvmChainsSettingsKey = "blockchain.evm_chains"

// EvmChainSettings is one EVM chain's RPC endpoint.
type EvmChainSettings struct {
	ChainId string `koanf:"-"`
	HttpUrl string `koanf:"http_url"`
}

// EvmChainsSettings holds the operator-configured EVM chains the node has an RPC
// for, keyed by chain id (e.g. "80002"). A plain chain id, not a CAIP-2/eip155
// network, and a string so non-EVM ids stay representable.
type EvmChainsSettings map[string]EvmChainSettings

func NewEvmChainsSettings(logger *zap.Logger, k *koanf.Koanf) *EvmChainsSettings {
	configByChainId := map[string]*EvmChainSettings{}
	if err := k.UnmarshalWithConf(EvmChainsSettingsKey, &configByChainId, koanf.UnmarshalConf{Tag: "koanf"}); err != nil {
		logger.Fatal("failed to unmarshal blockchain evm chains settings", zap.Error(err))
	}

	settings := EvmChainsSettings{}
	for chainId, chain := range configByChainId {
		chainId = strings.TrimSpace(chainId)
		if chain == nil {
			logger.Fatal("blockchain evm chain config is nil", zap.String("chainId", chainId))
		}
		httpUrl := strings.TrimSpace(chain.HttpUrl)
		if len(chainId) == 0 || len(httpUrl) == 0 {
			logger.Fatal("blockchain evm chain needs a chain id key and an http_url", zap.String("chainId", chainId))
		}
		settings[chainId] = EvmChainSettings{ChainId: chainId, HttpUrl: httpUrl}
	}

	if len(settings) == 0 {
		logger.Fatal("configure at least one chain under blockchain.evm_chains (or BLOCKCHAIN__EVM_CHAINS__<chainId>__HTTP_URL env)")
	}

	return &settings
}

// Has reports whether the node has an RPC for this chain.
func (s *EvmChainsSettings) Has(chainId string) bool {
	if s == nil {
		return false
	}
	_, ok := (*s)[chainId]
	return ok
}

// HttpUrlFor returns the RPC URL for a chain, if configured.
func (s *EvmChainsSettings) HttpUrlFor(chainId string) (string, bool) {
	if s == nil {
		return "", false
	}
	chain, ok := (*s)[chainId]
	return chain.HttpUrl, ok
}

// ChainIds lists the configured chain ids, sorted (map iteration is unordered).
func (s *EvmChainsSettings) ChainIds() []string {
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
