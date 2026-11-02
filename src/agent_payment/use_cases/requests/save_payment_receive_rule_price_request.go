package requests

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/6022-labs/agentic-agent-node/src/common/errors"
	"github.com/ethereum/go-ethereum/common"
	"github.com/holiman/uint256"
)

// SavePaymentReceiveRulePriceRequest is one price of a rule and the payers it applies to.
type SavePaymentReceiveRulePriceRequest struct {
	Price           string   `json:"price"`
	AddressPatterns []string `json:"addressPatterns"`
}

// ValidateAndSanitize canonicalizes the row; index names it in error fields.
func (a *SavePaymentReceiveRulePriceRequest) ValidateAndSanitize(index int) error {
	priceField := fmt.Sprintf("prices[%d].price", index)
	patternsField := fmt.Sprintf("prices[%d].addressPatterns", index)

	price, err := uint256.FromDecimal(strings.TrimSpace(a.Price))
	if err != nil {
		return errors.NewValidationError(
			priceField,
			"must be a non-negative integer in the asset's atomic units (0 = free)",
		)
	}
	// Canonical so "01" and "1" cannot be two rows of one price.
	a.Price = price.Dec()

	if len(a.AddressPatterns) == 0 {
		return errors.NewValidationError(
			patternsField,
			"at least one pattern is required (\".*\" opens it to everyone)",
		)
	}
	patterns := make([]string, 0, len(a.AddressPatterns))
	seen := make(map[string]bool, len(a.AddressPatterns))
	for _, pattern := range a.AddressPatterns {
		trimmed := strings.TrimSpace(pattern)
		if len(trimmed) == 0 {
			return errors.NewValidationError(patternsField, "patterns cannot be empty")
		}
		if _, err := regexp.Compile(trimmed); err != nil {
			return errors.NewValidationError(
				patternsField,
				"must be addresses or valid regexes",
			)
		}
		if common.IsHexAddress(trimmed) {
			trimmed = strings.ToLower(trimmed)
		}
		if seen[trimmed] {
			continue
		}
		seen[trimmed] = true
		patterns = append(patterns, trimmed)
	}
	// ".*" already covers every caller, so the others are noise.
	if seen[".*"] {
		patterns = []string{".*"}
	}
	a.AddressPatterns = patterns
	return nil
}
