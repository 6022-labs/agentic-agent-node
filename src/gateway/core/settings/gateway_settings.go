package settings

import (
	"strings"

	"github.com/knadh/koanf/v2"
	"go.uber.org/zap"
)

const GatewaySettingsKey = "gateway"

const defaultPublicPrefix = "/api"

// GatewaySettings holds public-exposure config; PublicPrefix is the nginx mount path used in advertised absolute URLs.
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
