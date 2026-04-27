package settings

import (
	"strings"

	"github.com/knadh/koanf/v2"
	"go.uber.org/zap"
)

const HostSettingsKey = "gateway_host"

type HostSettings struct {
	ApiPort        int    `koanf:"api_port"`
	AppName        string `koanf:"app_name"`
	BodyLimitBytes int    `koanf:"body_limit_bytes"`
}

func NewHostSettings(logger *zap.Logger, k *koanf.Koanf) *HostSettings {
	settings := HostSettings{
		ApiPort:        5000,
		BodyLimitBytes: 4 * 1024 * 1024,
	}
	if err := k.Unmarshal(HostSettingsKey, &settings); err != nil {
		logger.Fatal("failed to unmarshal host settings", zap.Error(err))
	}

	settings.AppName = strings.TrimSpace(settings.AppName)

	if settings.ApiPort <= 0 {
		logger.Warn("invalid gateway_host.api_port value, defaulting to 5000")
		settings.ApiPort = 5000
	}

	if len(settings.AppName) == 0 {
		logger.Fatal("please set your gateway_host.app_name value in your configuration")
	}

	if settings.BodyLimitBytes <= 0 {
		logger.Warn("invalid gateway_host.body_limit_bytes value, defaulting to 4MB")
		settings.BodyLimitBytes = 4 * 1024 * 1024
	}

	return &settings
}
