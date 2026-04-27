package settings

import (
	"strings"

	"github.com/knadh/koanf/v2"
	"go.uber.org/zap"
)

const HostSettingsKey = "runtime_host"

// HostSettings configures the runtime binary. ListenAddress defaults to
// "127.0.0.1" — runtime is reachable only by gateway_host on the loopback
// interface.
type HostSettings struct {
	ApiPort        int    `koanf:"api_port"`
	AppName        string `koanf:"app_name"`
	BodyLimitBytes int    `koanf:"body_limit_bytes"`
	ListenAddress  string `koanf:"listen_address"`
}

func NewHostSettings(logger *zap.Logger, k *koanf.Koanf) *HostSettings {
	settings := HostSettings{
		ApiPort:        5001,
		BodyLimitBytes: 4 * 1024 * 1024,
		ListenAddress:  "127.0.0.1",
	}
	if err := k.Unmarshal(HostSettingsKey, &settings); err != nil {
		logger.Fatal("failed to unmarshal runtime_host settings", zap.Error(err))
	}

	settings.AppName = strings.TrimSpace(settings.AppName)
	settings.ListenAddress = strings.TrimSpace(settings.ListenAddress)

	if settings.ApiPort <= 0 {
		logger.Warn("invalid runtime_host.api_port value, defaulting to 5001")
		settings.ApiPort = 5001
	}
	if len(settings.AppName) == 0 {
		logger.Fatal("please set your runtime_host.app_name value in your configuration")
	}
	if settings.BodyLimitBytes <= 0 {
		logger.Warn("invalid runtime_host.body_limit_bytes value, defaulting to 4MB")
		settings.BodyLimitBytes = 4 * 1024 * 1024
	}
	if len(settings.ListenAddress) == 0 {
		logger.Warn("invalid runtime_host.listen_address value, defaulting to 127.0.0.1")
		settings.ListenAddress = "127.0.0.1"
	}

	return &settings
}
