package settings

import (
	"github.com/knadh/koanf/v2"
	"go.uber.org/zap"
)

const EnsBlockchainSettingsKey = "blockchain"

// EnsBlockchainSettings is the ENS module's own copy of the chain id it signs
// and namehashes against. Deliberately duplicated from agent_identity_blockchain
// so the ENS module depends on no other module's settings — it reads the same
// BLOCKCHAIN__ID value.
type EnsBlockchainSettings struct {
	ChainId uint64 `koanf:"id"`
}

func NewEnsBlockchainSettings(logger *zap.Logger, k *koanf.Koanf) *EnsBlockchainSettings {
	var s EnsBlockchainSettings
	if err := k.Unmarshal(EnsBlockchainSettingsKey, &s); err != nil {
		logger.Fatal("failed to unmarshal blockchain settings for ENS", zap.Error(err))
	}

	if s.ChainId == 0 {
		logger.Fatal("please set blockchain.id (or BLOCKCHAIN__ID env)")
	}

	return &s
}
