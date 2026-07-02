package settings

import (
	_ "embed"
	"encoding/json"
	"strconv"
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

// Token6022Settings holds the $6022 ERC-20 address per chain id (inner join of
// the built-in registry and the configured chains). It rates an agent by its
// creator's balance.
type Token6022Settings map[string]common.Address

func NewToken6022Settings(logger *zap.Logger, evmChainsSettings *EvmChainsSettings) *Token6022Settings {
	var f token6022File
	if err := json.Unmarshal(token6022Json, &f); err != nil {
		logger.Fatal("failed to parse built-in token_6022.json", zap.Error(err))
	}

	settings := Token6022Settings{}
	for _, t := range f.Tokens {
		addr := strings.TrimSpace(t.Address)
		chainId := strconv.FormatUint(t.ChainId, 10)
		if len(addr) == 0 || !evmChainsSettings.Has(chainId) {
			continue
		}
		settings[chainId] = common.HexToAddress(addr)
	}
	return &settings
}

// AddressFor returns the $6022 token address for a chain, if any.
func (s *Token6022Settings) AddressFor(chainId string) (common.Address, bool) {
	if s == nil {
		return common.Address{}, false
	}
	addr, ok := (*s)[chainId]
	return addr, ok
}
