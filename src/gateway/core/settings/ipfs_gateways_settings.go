package settings

import (
	_ "embed"
	"encoding/json"
	"strings"

	"github.com/knadh/koanf/v2"
	"go.uber.org/zap"
)

const IPFSGatewaysSettingsKey = "ipfs"

//go:embed ipfs_gateways.json
var ipfsGatewaysJson []byte

type ipfsGatewaysFile struct {
	Templates []string `json:"templates"`
}

// IPFSGatewaysSettings lists the gateway URL templates the console tries in order.
type IPFSGatewaysSettings struct {
	GatewayUrlTemplates []string `koanf:"gateway_url_templates"`
}

func NewIPFSGatewaysSettings(logger *zap.Logger, k *koanf.Koanf) *IPFSGatewaysSettings {
	// Not seeded: mapstructure merges into an existing slice element-wise.
	var s IPFSGatewaysSettings
	if err := k.Unmarshal(IPFSGatewaysSettingsKey, &s); err != nil {
		logger.Fatal("failed to unmarshal ipfs settings", zap.Error(err))
	}

	s.normalizeTemplates()
	if len(s.GatewayUrlTemplates) == 0 {
		s.applyBuiltInTemplates(logger)
	}

	return &s
}

func (s *IPFSGatewaysSettings) applyBuiltInTemplates(logger *zap.Logger) {
	var file ipfsGatewaysFile
	if err := json.Unmarshal(ipfsGatewaysJson, &file); err != nil {
		logger.Fatal("failed to parse built-in ipfs_gateways.json", zap.Error(err))
	}

	s.GatewayUrlTemplates = file.Templates
	s.normalizeTemplates()
}

// Env hands the whole list through as one comma-separated string.
func (s *IPFSGatewaysSettings) normalizeTemplates() {
	normalized := make([]string, 0, len(s.GatewayUrlTemplates))
	for _, template := range s.GatewayUrlTemplates {
		for _, part := range strings.Split(template, ",") {
			if trimmed := strings.TrimSpace(part); len(trimmed) != 0 {
				normalized = append(normalized, trimmed)
			}
		}
	}

	s.GatewayUrlTemplates = normalized
}
