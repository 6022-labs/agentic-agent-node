package settings

import (
	"github.com/knadh/koanf/v2"
	"go.uber.org/zap"
)

const HostSettingsKey = "runtime_host"

type HostSettings struct {
	ApiPort        int `koanf:"api_port"`
	BodyLimitBytes int `koanf:"body_limit_bytes"`
}

func NewHostSettings(logger *zap.Logger, k *koanf.Koanf) *HostSettings {
	settings := HostSettings{
		ApiPort:        5400,
		BodyLimitBytes: 4 * 1024 * 1024,
	}
	if err := k.Unmarshal(HostSettingsKey, &settings); err != nil {
		logger.Fatal("failed to unmarshal runtime_host settings", zap.Error(err))
	}

	if settings.ApiPort <= 0 {
		logger.Warn("invalid runtime_host.api_port value, defaulting to 5400")
		settings.ApiPort = 5400
	}
	if settings.BodyLimitBytes <= 0 {
		logger.Warn("invalid runtime_host.body_limit_bytes value, defaulting to 4MB")
		settings.BodyLimitBytes = 4 * 1024 * 1024
	}

	return &settings
}
