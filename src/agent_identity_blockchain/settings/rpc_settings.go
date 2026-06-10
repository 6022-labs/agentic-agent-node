package settings

import (
	"strings"

	"github.com/knadh/koanf/v2"
	"go.uber.org/zap"
)

const RpcSettingsKey = "rpc"

// RpcSettings is the EVM RPC endpoint the adapter dials. Adapter-only: nothing
// outside agent_identity_blockchain needs the URL.
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
