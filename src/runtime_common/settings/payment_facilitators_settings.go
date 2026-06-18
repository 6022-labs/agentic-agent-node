package settings

import (
	"sort"
	"strconv"
	"strings"

	"github.com/6022-labs/agentic-agent-node/src/common/helpers"
	"github.com/knadh/koanf/v2"
	"go.uber.org/zap"
)

const PaymentFacilitatorsSettingsKey = "payment.facilitators"

// defaultMaxTimeoutSeconds is the x402 settlement window applied when a
// facilitator entry doesn't set one.
const defaultMaxTimeoutSeconds = 60

// PaymentFacilitator is one network's x402 facilitator: where to settle and how
// long callers have to do it. The timeout is network-level, shared by every
// rule on that network. Network is the map key, set on Entries() output.
type PaymentFacilitator struct {
	Network           string `koanf:"-"`
	BaseUrl           string `koanf:"base_url"`
	MaxTimeoutSeconds int    `koanf:"max_timeout_seconds"`
}

// PaymentFacilitatorsSettings holds the operator-configured x402 facilitators
// keyed by CAIP-2 network. Each must have a matching RPC in blockchain.chains,
// else its assets can't be validated and it could never settle.
type PaymentFacilitatorsSettings map[string]PaymentFacilitator

func NewPaymentFacilitatorsSettings(logger *zap.Logger, k *koanf.Koanf, chainsSettings *ChainsSettings) *PaymentFacilitatorsSettings {
	configByChainId := map[string]*PaymentFacilitator{}
	if err := k.UnmarshalWithConf(PaymentFacilitatorsSettingsKey, &configByChainId, koanf.UnmarshalConf{Tag: "koanf"}); err != nil {
		logger.Fatal("failed to unmarshal payment facilitators settings", zap.Error(err))
	}

	settings := PaymentFacilitatorsSettings{}
	for chainIdKey, facilitator := range configByChainId {
		chainIdKey = strings.TrimSpace(chainIdKey)
		if facilitator == nil {
			logger.Fatal("payment facilitator config is nil", zap.String("chainId", chainIdKey))
		}
		baseUrl := strings.TrimRight(strings.TrimSpace(facilitator.BaseUrl), "/")
		if len(chainIdKey) == 0 || len(baseUrl) == 0 {
			logger.Fatal("payment facilitator needs a chain id key and base_url", zap.String("chainId", chainIdKey))
		}

		// Config keys facilitators by bare chain id to mirror blockchain.chains;
		// the payment domain matches on CAIP-2, so the stored key is reconstructed.
		chainId, err := strconv.ParseUint(chainIdKey, 10, 64)
		if err != nil || chainId == 0 {
			logger.Fatal("payment facilitator key must be a numeric chain id, e.g. 80002", zap.String("chainId", chainIdKey))
		}
		if !chainsSettings.Has(chainIdKey) {
			logger.Fatal("payment facilitator has no matching RPC in blockchain.chains; add the chain or drop the facilitator",
				zap.String("chainId", chainIdKey),
			)
		}

		maxTimeoutSeconds := facilitator.MaxTimeoutSeconds
		if maxTimeoutSeconds <= 0 {
			maxTimeoutSeconds = defaultMaxTimeoutSeconds
		}
		settings[helpers.FormatCaip2(chainId)] = PaymentFacilitator{BaseUrl: baseUrl, MaxTimeoutSeconds: maxTimeoutSeconds}
	}

	return &settings
}

// BaseUrlFor returns the facilitator base URL for a network, if configured.
func (s *PaymentFacilitatorsSettings) BaseUrlFor(network string) (string, bool) {
	if s == nil {
		return "", false
	}
	facilitator, ok := (*s)[network]
	return facilitator.BaseUrl, ok
}

// Networks lists the configured networks, sorted (map iteration is unordered).
func (s *PaymentFacilitatorsSettings) Networks() []string {
	if s == nil {
		return nil
	}
	networks := make([]string, 0, len(*s))
	for network := range *s {
		networks = append(networks, network)
	}
	sort.Strings(networks)
	return networks
}

// Entries lists the facilitators in sorted-network order, each tagged with its
// network (the map key).
func (s *PaymentFacilitatorsSettings) Entries() []PaymentFacilitator {
	networks := s.Networks()
	entries := make([]PaymentFacilitator, len(networks))
	for i, network := range networks {
		facilitator := (*s)[network]
		facilitator.Network = network
		entries[i] = facilitator
	}
	return entries
}
