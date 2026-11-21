package settings

import (
	"strings"

	"github.com/btcsuite/btcd/chaincfg/v2"
	"github.com/knadh/koanf/v2"
	"go.uber.org/zap"
)

const BtcNetworkSettingsKey = "blockchain.btc_chains.network"

const (
	btcNetworkMainnet = "mainnet"
	btcNetworkTestnet = "testnet"

	// BtcNetworkIdMainnet/Testnet are the CAIP-2-style routing keys the bitcoin network is identified by.
	BtcNetworkIdMainnet = "bitcoin:mainnet"
	BtcNetworkIdTestnet = "bitcoin:testnet"
)

// BtcChainsSettings selects the bitcoin network the agent's BTC wallet lives on (one per node, default mainnet).
type BtcChainsSettings struct {
	network string
}

func NewBtcChainsSettings(logger *zap.Logger, k *koanf.Koanf) *BtcChainsSettings {
	network := strings.ToLower(strings.TrimSpace(k.String(BtcNetworkSettingsKey)))
	if len(network) == 0 {
		network = btcNetworkMainnet
	}
	if network != btcNetworkMainnet && network != btcNetworkTestnet {
		logger.Fatal("blockchain.btc_chains.network must be \"mainnet\" or \"testnet\"", zap.String("network", network))
	}
	return &BtcChainsSettings{network: network}
}

// ChainParams returns the btcsuite params used to derive and validate addresses.
func (s *BtcChainsSettings) ChainParams() *chaincfg.Params {
	if s.network == btcNetworkTestnet {
		return &chaincfg.TestNet3Params
	}
	return &chaincfg.MainNetParams
}

// NetworkId returns the bitcoin network routing key (bitcoin:mainnet/testnet).
func (s *BtcChainsSettings) NetworkId() string {
	if s.network == btcNetworkTestnet {
		return BtcNetworkIdTestnet
	}
	return BtcNetworkIdMainnet
}
