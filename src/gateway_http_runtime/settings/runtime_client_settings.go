package settings

import (
	"strings"

	"github.com/knadh/koanf/v2"
	"go.uber.org/zap"
)

const RuntimeClientSettingsKey = "runtime_client"

type RuntimeClientSettings struct {
	BaseUrl string
}

type runtimeClientSettingsRaw struct {
	BaseUrl string `koanf:"base_url"`
}

func NewRuntimeClientSettings(logger *zap.Logger, k *koanf.Koanf) *RuntimeClientSettings {
	raw := runtimeClientSettingsRaw{
		BaseUrl: "http://127.0.0.1:5400",
	}
	if err := k.Unmarshal(RuntimeClientSettingsKey, &raw); err != nil {
		logger.Fatal("failed to unmarshal runtime_client settings", zap.Error(err))
	}

	raw.BaseUrl = strings.TrimSpace(raw.BaseUrl)
	if len(raw.BaseUrl) == 0 {
		raw.BaseUrl = "http://127.0.0.1:5400"
	}

	return &RuntimeClientSettings{
		BaseUrl: raw.BaseUrl,
	}
}
