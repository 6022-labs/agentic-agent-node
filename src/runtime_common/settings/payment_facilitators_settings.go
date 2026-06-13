package settings

import (
	"sort"
	"strings"

	"github.com/knadh/koanf/v2"
	"go.uber.org/zap"
)

const PaymentFacilitatorsSettingsKey = "payment.facilitators"

// PaymentFacilitatorConfig is one network's x402 facilitator.
type PaymentFacilitatorConfig struct {
	BaseUrl string `koanf:"base_url"`
}

// PaymentFacilitatorEntry pairs a network with its facilitator base URL.
type PaymentFacilitatorEntry struct {
	Network string
	BaseUrl string
}

// PaymentFacilitatorsSettings holds the operator-configured x402 facilitators
// keyed by network. Paid rules may only target a configured network.
type PaymentFacilitatorsSettings map[string]string

func NewPaymentFacilitatorsSettings(logger *zap.Logger, k *koanf.Koanf) *PaymentFacilitatorsSettings {
	configByNetwork := map[string]*PaymentFacilitatorConfig{}
	if err := k.UnmarshalWithConf(PaymentFacilitatorsSettingsKey, &configByNetwork, koanf.UnmarshalConf{Tag: "koanf"}); err != nil {
		logger.Fatal("failed to unmarshal payment facilitators settings", zap.Error(err))
	}

	settings := PaymentFacilitatorsSettings{}
	for network, config := range configByNetwork {
		network = strings.TrimSpace(network)
		if config == nil {
			logger.Fatal("payment facilitator config is nil", zap.String("network", network))
		}
		baseUrl := strings.TrimRight(strings.TrimSpace(config.BaseUrl), "/")
		if len(network) == 0 || len(baseUrl) == 0 {
			logger.Fatal("payment facilitator needs a network key and base_url", zap.String("network", network))
		}
		settings[network] = baseUrl
	}

	return &settings
}

// BaseUrlFor returns the facilitator base URL for a network, if configured.
func (s *PaymentFacilitatorsSettings) BaseUrlFor(network string) (string, bool) {
	if s == nil {
		return "", false
	}
	baseUrl, ok := (*s)[network]
	return baseUrl, ok
}

// Networks lists the configured networks in sorted order (the map has none).
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

// Entries lists the configured facilitators in sorted-network order.
func (s *PaymentFacilitatorsSettings) Entries() []PaymentFacilitatorEntry {
	networks := s.Networks()
	entries := make([]PaymentFacilitatorEntry, len(networks))
	for i, network := range networks {
		entries[i] = PaymentFacilitatorEntry{Network: network, BaseUrl: (*s)[network]}
	}
	return entries
}
