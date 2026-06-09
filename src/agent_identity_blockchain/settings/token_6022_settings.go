package settings

import (
	_ "embed"
	"encoding/json"
	"strings"

	"github.com/ethereum/go-ethereum/common"
	"go.uber.org/zap"
)

//go:embed token_6022.json
var token6022Json []byte

type token6022Entry struct {
	ChainId uint64 `json:"chainId"`
	Name    string `json:"name"`
	Address string `json:"address"`
}

type token6022File struct {
	Tokens []token6022Entry `json:"tokens"`
}

// Token6022Settings holds the $6022 ERC-20 address for this agent's chain,
// bound at startup from BlockchainSettings.Id. It rates an agent by its
// creator's balance.
type Token6022Settings struct {
	SmartContractAddress common.Address
}

func NewToken6022Settings(
	logger *zap.Logger,
	blockchainSettings *BlockchainSettings,
) *Token6022Settings {
	var f token6022File
	if err := json.Unmarshal(token6022Json, &f); err != nil {
		logger.Fatal("failed to parse built-in token_6022.json", zap.Error(err))
	}

	for _, t := range f.Tokens {
		if t.ChainId != blockchainSettings.Id {
			continue
		}
		addr := strings.TrimSpace(t.Address)
		if len(addr) == 0 {
			logger.Fatal(
				"$6022 token address is empty for the configured chain; populate token_6022.json and rebuild",
				zap.Uint64("chainId", blockchainSettings.Id),
				zap.String("name", t.Name),
			)
		}
		logger.Info("resolved $6022 token from built-in registry",
			zap.Uint64("chainId", blockchainSettings.Id),
			zap.String("name", t.Name),
			zap.String("address", addr),
		)
		return &Token6022Settings{SmartContractAddress: common.HexToAddress(addr)}
	}

	logger.Fatal(
		"no $6022 token entry for the configured chain; add an entry to token_6022.json and rebuild",
		zap.Uint64("chainId", blockchainSettings.Id),
	)
	return nil
}
