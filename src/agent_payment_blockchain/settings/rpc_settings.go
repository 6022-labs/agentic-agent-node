package settings

import (
	"strings"

	"github.com/knadh/koanf/v2"
	"go.uber.org/zap"
)

const RpcSettingsKey = "rpc"

// RpcSettings is the EVM RPC endpoint this adapter dials. Adapter-only: each
// blockchain adapter owns its own RPC config and client so they stay decoupled.
type RpcSettings struct {
	HttpUrl string `koanf:"http_url"`
}

func NewRpcSettings(logger *zap.Logger, k *koanf.Koanf) *RpcSettings {
	var settings RpcSettings
	if err := k.Unmarshal(RpcSettingsKey, &settings); err != nil {
		logger.Fatal("failed to unmarshal rpc settings", zap.Error(err))
	}

	settings.HttpUrl = strings.TrimSpace(settings.HttpUrl)

	if len(settings.HttpUrl) == 0 {
		logger.Fatal("please set rpc.http_url (or RPC__HTTP_URL env)")
	}

	return &settings
}
