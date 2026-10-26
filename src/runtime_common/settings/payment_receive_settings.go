package settings

import (
	"sort"
	"strconv"
	"strings"

	"github.com/6022-labs/agentic-agent-node/src/common/helpers"
	"github.com/knadh/koanf/v2"
	"go.uber.org/zap"
)

const PaymentReceiveSettingsKey = "payment.receive.facilitators"

// legacyPaymentReceiveSettingsKey is refused loudly; keeping it would silently settle nothing.
const legacyPaymentReceiveSettingsKey = "payment.facilitators"

// defaultMaxTimeoutSeconds is the settlement window when an entry sets none.
const defaultMaxTimeoutSeconds = 60

// PaymentFacilitator is one network's x402 facilitator; the timeout is network-level.
type PaymentFacilitator struct {
	Network           string `koanf:"-"`
	BaseUrl           string `koanf:"base_url"`
	MaxTimeoutSeconds int    `koanf:"max_timeout_seconds"`
}

// PaymentReceiveSettings holds the x402 facilitators by CAIP-2 network; each needs an RPC in blockchain.evm_chains.
type PaymentReceiveSettings map[string]PaymentFacilitator

func NewPaymentReceiveSettings(
	logger *zap.Logger,
	k *koanf.Koanf,
	evmChainsSettings *EvmChainsSettings,
) *PaymentReceiveSettings {
	if k.Exists(legacyPaymentReceiveSettingsKey) {
		logger.Fatal(
			"payment.facilitators moved to payment.receive.facilitators; " +
				"rename PAYMENT__FACILITATORS__* to PAYMENT__RECEIVE__FACILITATORS__*",
		)
	}

	configByChainId := map[string]*PaymentFacilitator{}
	if err := k.UnmarshalWithConf(
		PaymentReceiveSettingsKey,
		&configByChainId,
		koanf.UnmarshalConf{Tag: "koanf"},
	); err != nil {
		logger.Fatal("failed to unmarshal payment facilitators settings", zap.Error(err))
	}

	settings := PaymentReceiveSettings{}
	for chainIdKey, facilitator := range configByChainId {
		chainIdKey = strings.TrimSpace(chainIdKey)
		if facilitator == nil {
			logger.Fatal("payment facilitator config is nil", zap.String("chainId", chainIdKey))
		}
		baseUrl := strings.TrimRight(strings.TrimSpace(facilitator.BaseUrl), "/")
		if len(chainIdKey) == 0 || len(baseUrl) == 0 {
			logger.Fatal("payment facilitator needs a chain id key and base_url", zap.String("chainId", chainIdKey))
		}

		// Config keys by bare chain id like blockchain.evm_chains; the domain matches on CAIP-2.
		chainId, err := strconv.ParseUint(chainIdKey, 10, 64)
		if err != nil || chainId == 0 {
			logger.Fatal(
				"payment facilitator key must be a numeric chain id, e.g. 80002",
				zap.String("chainId", chainIdKey),
			)
		}
		if !evmChainsSettings.Has(chainIdKey) {
			logger.Fatal(
				"payment facilitator has no matching RPC in blockchain.evm_chains; add the chain or drop the facilitator",
				zap.String("chainId", chainIdKey),
			)
		}

		maxTimeoutSeconds := facilitator.MaxTimeoutSeconds
		if maxTimeoutSeconds <= 0 {
			maxTimeoutSeconds = defaultMaxTimeoutSeconds
		}
		settings[helpers.FormatCaip2(chainId)] = PaymentFacilitator{
			BaseUrl:           baseUrl,
			MaxTimeoutSeconds: maxTimeoutSeconds,
		}
	}

	return &settings
}

// BaseUrlFor returns the facilitator base URL for a network, if configured.
func (s *PaymentReceiveSettings) BaseUrlFor(network string) (string, bool) {
	if s == nil {
		return "", false
	}
	facilitator, ok := (*s)[network]
	return facilitator.BaseUrl, ok
}

// Networks lists the configured networks, sorted (map iteration is unordered).
func (s *PaymentReceiveSettings) Networks() []string {
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

// Entries lists the facilitators in sorted-network order, each tagged with its network.
func (s *PaymentReceiveSettings) Entries() []PaymentFacilitator {
	networks := s.Networks()
	entries := make([]PaymentFacilitator, len(networks))
	for i, network := range networks {
		facilitator := (*s)[network]
		facilitator.Network = network
		entries[i] = facilitator
	}
	return entries
}
