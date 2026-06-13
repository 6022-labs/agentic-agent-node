package settings

import (
	"sort"
	"strings"

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
// keyed by network. Paid rules may only target a configured network.
type PaymentFacilitatorsSettings map[string]PaymentFacilitator

func NewPaymentFacilitatorsSettings(logger *zap.Logger, k *koanf.Koanf) *PaymentFacilitatorsSettings {
	configByNetwork := map[string]*PaymentFacilitator{}
	if err := k.UnmarshalWithConf(PaymentFacilitatorsSettingsKey, &configByNetwork, koanf.UnmarshalConf{Tag: "koanf"}); err != nil {
		logger.Fatal("failed to unmarshal payment facilitators settings", zap.Error(err))
	}

	settings := PaymentFacilitatorsSettings{}
	for network, facilitator := range configByNetwork {
		network = strings.TrimSpace(network)
		if facilitator == nil {
			logger.Fatal("payment facilitator config is nil", zap.String("network", network))
		}
		baseUrl := strings.TrimRight(strings.TrimSpace(facilitator.BaseUrl), "/")
		if len(network) == 0 || len(baseUrl) == 0 {
			logger.Fatal("payment facilitator needs a network key and base_url", zap.String("network", network))
		}
		maxTimeoutSeconds := facilitator.MaxTimeoutSeconds
		if maxTimeoutSeconds <= 0 {
			maxTimeoutSeconds = defaultMaxTimeoutSeconds
		}
		settings[network] = PaymentFacilitator{BaseUrl: baseUrl, MaxTimeoutSeconds: maxTimeoutSeconds}
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
