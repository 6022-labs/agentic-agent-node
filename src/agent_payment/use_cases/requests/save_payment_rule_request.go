package requests

import (
	"math/big"
	"regexp"
	"strconv"
	"strings"

	"github.com/6022-labs/agentic-agent-node/src/common/errors"
	"github.com/ethereum/go-ethereum/common"
)

const (
	// EveryoneAddressPattern opens a rule to everyone; only such rules are
	// advertised in anonymous 402 responses.
	EveryoneAddressPattern   = ".*"
	caip2Eip155Prefix        = "eip155:"
	defaultMaxTimeoutSeconds = 60
	defaultEip712Name        = "USDC"
	defaultEip712Version     = "2"
)

type SavePaymentRuleRequest struct {
	Name              string   `json:"name"`
	Network           string   `json:"network"`
	Asset             string   `json:"asset"`
	PayTo             string   `json:"payTo"`
	Price             string   `json:"price"`
	MaxTimeoutSeconds int      `json:"maxTimeoutSeconds,omitempty"`
	Eip712Name        string   `json:"eip712Name,omitempty"`
	Eip712Version     string   `json:"eip712Version,omitempty"`
	AddressPatterns   []string `json:"addressPatterns"`
}

// ValidateAndSanitize trims, applies defaults and rejects malformed values.
func (a *SavePaymentRuleRequest) ValidateAndSanitize() error {
	a.Name = strings.TrimSpace(a.Name)
	a.Network = strings.TrimSpace(a.Network)
	a.Asset = strings.TrimSpace(a.Asset)
	a.PayTo = strings.TrimSpace(a.PayTo)
	a.Price = strings.TrimSpace(a.Price)
	a.Eip712Name = strings.TrimSpace(a.Eip712Name)
	a.Eip712Version = strings.TrimSpace(a.Eip712Version)

	if len(a.Name) == 0 || len(a.Name) > 64 {
		return errors.NewValidationError("name", "must be 1-64 characters")
	}
	if _, err := ParseEip155ChainId(a.Network); err != nil {
		return errors.NewValidationError("network", "must be a CAIP-2 eip155 identifier, e.g. eip155:80002")
	}
	if !common.IsHexAddress(a.Asset) {
		return errors.NewValidationError("asset", "must be a hex address")
	}
	if !common.IsHexAddress(a.PayTo) {
		return errors.NewValidationError("payTo", "must be a hex address")
	}
	price, ok := new(big.Int).SetString(a.Price, 10)
	if !ok || price.Sign() < 0 {
		return errors.NewValidationError("price", "must be a non-negative integer in the asset's atomic units (0 = free)")
	}
	if len(a.AddressPatterns) == 0 {
		return errors.NewValidationError("addressPatterns", "at least one pattern is required (\".*\" opens the rule to everyone)")
	}
	for i, pattern := range a.AddressPatterns {
		trimmed := strings.TrimSpace(pattern)
		if len(trimmed) == 0 {
			return errors.NewValidationError("addressPatterns", "patterns cannot be empty")
		}
		if _, err := regexp.Compile(trimmed); err != nil {
			return errors.NewValidationError("addressPatterns", "must be valid regexes matched against the lowercase 0x address")
		}
		a.AddressPatterns[i] = trimmed
	}

	if a.MaxTimeoutSeconds <= 0 {
		a.MaxTimeoutSeconds = defaultMaxTimeoutSeconds
	}
	if len(a.Eip712Name) == 0 {
		a.Eip712Name = defaultEip712Name
	}
	if len(a.Eip712Version) == 0 {
		a.Eip712Version = defaultEip712Version
	}
	return nil
}

// ParseEip155ChainId extracts the chain id from a CAIP-2 eip155 network
// identifier ("eip155:80002" → 80002).
func ParseEip155ChainId(network string) (int64, error) {
	rest, found := strings.CutPrefix(network, caip2Eip155Prefix)
	if !found {
		return 0, errors.NewValidationError("network", "missing eip155 prefix")
	}
	chainId, err := strconv.ParseInt(rest, 10, 64)
	if err != nil || chainId <= 0 {
		return 0, errors.NewValidationError("network", "invalid chain id")
	}
	return chainId, nil
}
