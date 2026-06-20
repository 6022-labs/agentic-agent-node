package settings

import (
	"strings"

	"github.com/knadh/koanf/v2"
	"go.uber.org/zap"
)

const TelemetrySettingsKey = "telemetry"

type TelemetrySettings struct {
	Enabled     bool   `koanf:"enabled"`
	Endpoint    string `koanf:"endpoint"`
	Insecure    bool   `koanf:"insecure"`
	ServiceName string `koanf:"service_name"`
	NodeId      string `koanf:"node_id"`
}

func NewTelemetrySettings(
	logger *zap.Logger,
	k *koanf.Koanf,
) *TelemetrySettings {
	logger.Debug("loading settings", zap.String("key", TelemetrySettingsKey))

	var settings TelemetrySettings
	if err := k.UnmarshalWithConf(TelemetrySettingsKey, &settings, koanf.UnmarshalConf{Tag: "koanf"}); err != nil {
		logger.Fatal("failed to unmarshal telemetry settings", zap.Error(err))
	}

	settings.Endpoint = strings.TrimSpace(settings.Endpoint)
	settings.ServiceName = strings.TrimSpace(settings.ServiceName)
	settings.NodeId = strings.TrimSpace(settings.NodeId)

	// service_name is not required: each host supplies its own default in code (shared container/env).
	if settings.Enabled {
		if settings.Endpoint == "" {
			logger.Fatal("telemetry.endpoint is required when telemetry.enabled is true")
		}
		if settings.NodeId == "" {
			logger.Fatal("telemetry.node_id is required when telemetry.enabled is true")
		}
	}

	logger.Debug("loaded settings", zap.String("key", TelemetrySettingsKey))
	return &settings
}
