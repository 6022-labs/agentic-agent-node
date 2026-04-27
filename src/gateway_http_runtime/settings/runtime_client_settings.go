package settings

import (
	"strings"
	"time"

	"github.com/knadh/koanf/v2"
	"go.uber.org/zap"
)

const RuntimeClientSettingsKey = "runtime_client"

type RuntimeClientSettings struct {
	BaseUrl string        `koanf:"base_url"`
	Timeout time.Duration `koanf:"timeout"`
}

func NewRuntimeClientSettings(logger *zap.Logger, k *koanf.Koanf) *RuntimeClientSettings {
	settings := RuntimeClientSettings{
		BaseUrl: "http://127.0.0.1:5001",
		Timeout: 5 * time.Second,
	}
	if err := k.Unmarshal(RuntimeClientSettingsKey, &settings); err != nil {
		logger.Fatal("failed to unmarshal runtime_client settings", zap.Error(err))
	}

	settings.BaseUrl = strings.TrimSpace(settings.BaseUrl)
	if len(settings.BaseUrl) == 0 {
		settings.BaseUrl = "http://127.0.0.1:5001"
	}
	if settings.Timeout <= 0 {
		settings.Timeout = 5 * time.Second
	}
	return &settings
}
