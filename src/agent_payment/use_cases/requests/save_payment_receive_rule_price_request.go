package requests

import (
	"fmt"
	"strings"

	payment_helpers "github.com/6022-labs/agentic-agent-node/src/agent_payment/helpers"
	"github.com/6022-labs/agentic-agent-node/src/common/errors"
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
	patterns, err := payment_helpers.SanitizeAddressPatterns(patternsField, a.AddressPatterns)
	if err != nil {
		return err
	}
	a.AddressPatterns = patterns
	return nil
}
