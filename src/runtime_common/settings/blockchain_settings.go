package settings

import (
	"github.com/knadh/koanf/v2"
	"go.uber.org/zap"
)

const BlockchainSettingsKey = "blockchain"

// BlockchainSettings is the chain the node is bound to. The RPC endpoint is
// adapter-only and lives in agent_identity_blockchain/settings.RpcSettings.
type BlockchainSettings struct {
	Id uint64 `koanf:"id"`
}

func NewBlockchainSettings(logger *zap.Logger, k *koanf.Koanf) *BlockchainSettings {
	var settings BlockchainSettings
	if err := k.Unmarshal(BlockchainSettingsKey, &settings); err != nil {
		logger.Fatal("failed to unmarshal blockchain settings", zap.Error(err))
	}

	if settings.Id == 0 {
		logger.Fatal("please set blockchain.id (or BLOCKCHAIN__ID env)")
	}

	return &settings
}
