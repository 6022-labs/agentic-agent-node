package settings

import (
	"strings"
	"time"

	"github.com/knadh/koanf/v2"
	"go.uber.org/zap"
)

const RuntimeClientSettingsKey = "runtime_client"

type RuntimeClientSettings struct {
	BaseUrl string
	Timeout time.Duration
}

type runtimeClientSettingsRaw struct {
	BaseUrl        string `koanf:"base_url"`
	TimeoutSeconds int    `koanf:"timeout"`
}

func NewRuntimeClientSettings(logger *zap.Logger, k *koanf.Koanf) *RuntimeClientSettings {
	raw := runtimeClientSettingsRaw{
		BaseUrl:        "http://127.0.0.1:5001",
		TimeoutSeconds: 5,
	}
	if err := k.Unmarshal(RuntimeClientSettingsKey, &raw); err != nil {
		logger.Fatal("failed to unmarshal runtime_client settings", zap.Error(err))
	}

	raw.BaseUrl = strings.TrimSpace(raw.BaseUrl)
	if len(raw.BaseUrl) == 0 {
		raw.BaseUrl = "http://127.0.0.1:5001"
	}
	if raw.TimeoutSeconds <= 0 {
		logger.Warn("invalid runtime_client.timeout value, defaulting to 5")
		raw.TimeoutSeconds = 5
	}

	return &RuntimeClientSettings{
		BaseUrl: raw.BaseUrl,
		Timeout: time.Duration(raw.TimeoutSeconds) * time.Second,
	}
}
