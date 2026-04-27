package settings

import (
	"strings"

	"github.com/knadh/koanf/v2"
	"go.uber.org/zap"
)

const BlockchainSettingsKey = "blockchain"

type BlockchainSettings struct {
	Id         uint64 `koanf:"id"`
	RpcHttpUrl string `koanf:"rpc_http_url"`
}

func NewBlockchainSettings(logger *zap.Logger, k *koanf.Koanf) *BlockchainSettings {
	var settings BlockchainSettings
	if err := k.Unmarshal(BlockchainSettingsKey, &settings); err != nil {
		logger.Fatal("failed to unmarshal blockchain settings", zap.Error(err))
	}

	settings.RpcHttpUrl = strings.TrimSpace(settings.RpcHttpUrl)

	if settings.Id == 0 {
		logger.Fatal("please set blockchain.id (or BLOCKCHAIN__ID env)")
	}
	if len(settings.RpcHttpUrl) == 0 {
		logger.Fatal("please set blockchain.rpc_http_url (or BLOCKCHAIN__RPC_HTTP_URL env)")
	}

	return &settings
}
