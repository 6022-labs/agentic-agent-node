package settings

import (
	"strings"

	"github.com/knadh/koanf/v2"
	"go.uber.org/zap"
)

const GatewaySettingsKey = "gateway"

const defaultPublicPrefix = "/api"

// GatewaySettings holds gateway-wide public-exposure config; PublicPrefix is the reverse-proxy mount path (nginx) the gateway is served under, used to build any absolute URL it advertises.
type GatewaySettings struct {
	PublicPrefix string `koanf:"public_prefix"`
}

func NewGatewaySettings(logger *zap.Logger, k *koanf.Koanf) *GatewaySettings {
	s := GatewaySettings{PublicPrefix: defaultPublicPrefix}
	if err := k.Unmarshal(GatewaySettingsKey, &s); err != nil {
		logger.Fatal("failed to unmarshal gateway settings", zap.Error(err))
	}

	trimmed := strings.Trim(strings.TrimSpace(s.PublicPrefix), "/")
	s.PublicPrefix = "/" + trimmed

	return &s
}
