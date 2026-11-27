package settings

import (
	"github.com/knadh/koanf/v2"
	"go.uber.org/zap"
)

const A2ASettingsKey = "a2a"

// A2ASettings bounds an outbound call; RequestTimeoutSeconds stays under the inference loop ceiling.
type A2ASettings struct {
	RequestTimeoutSeconds int `koanf:"request_timeout_seconds"`
	// CardTimeoutSeconds keeps a hanging card fetch from eating the turn's budget before the call itself.
	CardTimeoutSeconds int   `koanf:"card_timeout_seconds"`
	MaxResponseBytes   int64 `koanf:"max_response_bytes"`
	// AllowPrivateNetworks lets calls reach loopback, private addresses and plain http, for a local network.
	AllowPrivateNetworks bool `koanf:"allow_private_networks"`
}

func NewA2ASettings(logger *zap.Logger, k *koanf.Koanf) *A2ASettings {
	settings := A2ASettings{
		RequestTimeoutSeconds: 120,
		CardTimeoutSeconds:    10,
		MaxResponseBytes:      64 << 10,
		AllowPrivateNetworks:  false,
	}
	if err := k.Unmarshal(A2ASettingsKey, &settings); err != nil {
		logger.Fatal("failed to unmarshal a2a settings", zap.Error(err))
	}

	if settings.RequestTimeoutSeconds <= 0 {
		logger.Warn("invalid a2a.request_timeout_seconds, defaulting to 120")
		settings.RequestTimeoutSeconds = 120
	}
	if settings.CardTimeoutSeconds <= 0 {
		logger.Warn("invalid a2a.card_timeout_seconds, defaulting to 10")
		settings.CardTimeoutSeconds = 10
	}
	if settings.MaxResponseBytes <= 0 {
		logger.Warn("invalid a2a.max_response_bytes, defaulting to 65536")
		settings.MaxResponseBytes = 64 << 10
	}

	return &settings
}
