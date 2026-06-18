package settings

import (
	"slices"
	"strings"

	"github.com/knadh/koanf/v2"
	"go.uber.org/zap"
)

const ChainsSettingsKey = "blockchain.chains"

// ChainSettings is one chain's RPC endpoint.
type ChainSettings struct {
	ChainId string `koanf:"-"`
	HttpUrl string `koanf:"http_url"`
}

// ChainsSettings holds the operator-configured chains the node has an RPC for,
// keyed by chain id (e.g. "80002"). A plain chain id, not a CAIP-2/eip155
// network, and a string so non-EVM ids stay representable.
type ChainsSettings map[string]ChainSettings

func NewChainsSettings(logger *zap.Logger, k *koanf.Koanf) *ChainsSettings {
	configByChainId := map[string]*ChainSettings{}
	if err := k.UnmarshalWithConf(ChainsSettingsKey, &configByChainId, koanf.UnmarshalConf{Tag: "koanf"}); err != nil {
		logger.Fatal("failed to unmarshal blockchain chains settings", zap.Error(err))
	}

	settings := ChainsSettings{}
	for chainId, chain := range configByChainId {
		chainId = strings.TrimSpace(chainId)
		if chain == nil {
			logger.Fatal("blockchain chain config is nil", zap.String("chainId", chainId))
		}
		httpUrl := strings.TrimSpace(chain.HttpUrl)
		if len(chainId) == 0 || len(httpUrl) == 0 {
			logger.Fatal("blockchain chain needs a chain id key and an http_url", zap.String("chainId", chainId))
		}
		settings[chainId] = ChainSettings{ChainId: chainId, HttpUrl: httpUrl}
	}

	if len(settings) == 0 {
		logger.Fatal("configure at least one chain under blockchain.chains (or BLOCKCHAIN__CHAINS__<chainId>__HTTP_URL env)")
	}

	return &settings
}

// Has reports whether the node has an RPC for this chain.
func (s *ChainsSettings) Has(chainId string) bool {
	if s == nil {
		return false
	}
	_, ok := (*s)[chainId]
	return ok
}

// HttpUrlFor returns the RPC URL for a chain, if configured.
func (s *ChainsSettings) HttpUrlFor(chainId string) (string, bool) {
	if s == nil {
		return "", false
	}
	chain, ok := (*s)[chainId]
	return chain.HttpUrl, ok
}

// ChainIds lists the configured chain ids, sorted (map iteration is unordered).
func (s *ChainsSettings) ChainIds() []string {
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
