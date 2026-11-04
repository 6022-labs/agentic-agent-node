package requests

import (
	"encoding/json"
	"fmt"
	"math/big"
	"strings"

	"github.com/6022-labs/agentic-agent-node/src/agent_payment/helpers"
	"github.com/6022-labs/agentic-agent-node/src/common/errors"
)

// SavePaymentSpendRuleRequest is method-neutral: the limits, name and method are generic.
type SavePaymentSpendRuleRequest struct {
	SavePaymentRuleRequest
	MaxAmountPerPayment *string                             `json:"maxAmountPerPayment,omitempty"`
	Windows             []SavePaymentSpendRuleWindowRequest `json:"windows"`
	Config              json.RawMessage                     `json:"config"                        swaggertype:"object"`
}

func (a *SavePaymentSpendRuleRequest) ValidateAndSanitize() error {
	if err := a.SavePaymentRuleRequest.ValidateAndSanitize(); err != nil {
		return err
	}

	maxAmountPerPayment, err := a.validateMaxAmountPerPayment()
	if err != nil {
		return err
	}
	return a.validateWindows(maxAmountPerPayment)
}

// validateMaxAmountPerPayment treats a blank cap as none; "0" denies every non-zero quote.
func (a *SavePaymentSpendRuleRequest) validateMaxAmountPerPayment() (*big.Int, error) {
	if a.MaxAmountPerPayment == nil || len(strings.TrimSpace(*a.MaxAmountPerPayment)) == 0 {
		a.MaxAmountPerPayment = nil
		return nil, nil
	}

	maxAmountPerPayment, err := helpers.ParseSpendAmount("maxAmountPerPayment", *a.MaxAmountPerPayment)
	if err != nil {
		return nil, err
	}
	// Canonical so "007" is stored as "7", like receive prices.
	canonical := maxAmountPerPayment.Dec()
	a.MaxAmountPerPayment = &canonical
	return maxAmountPerPayment.ToBig(), nil
}

// validateWindows rejects a rule with no limit at all, a repeated period, and a cap no window could fit.
func (a *SavePaymentSpendRuleRequest) validateWindows(maxAmountPerPayment *big.Int) error {
	if maxAmountPerPayment == nil && len(a.Windows) == 0 {
		return errors.NewValidationError("windows", "a rule needs maxAmountPerPayment, a window, or both")
	}
	// Every signature locks and updates each window, so the count stays small.
	if len(a.Windows) > MaxPaymentSpendRuleWindows {
		return errors.NewValidationError("windows", fmt.Sprintf("at most %d windows", MaxPaymentSpendRuleWindows))
	}

	seen := make(map[int]bool, len(a.Windows))
	for i := range a.Windows {
		maxAmount, err := a.Windows[i].ValidateAndSanitize(i)
		if err != nil {
			return err
		}
		if seen[a.Windows[i].WindowSeconds] {
			return errors.NewValidationError("windows", "each windowSeconds may appear only once")
		}
		seen[a.Windows[i].WindowSeconds] = true

		if maxAmountPerPayment != nil && maxAmountPerPayment.Cmp(maxAmount) > 0 {
			return errors.NewValidationError(
				"maxAmountPerPayment",
				"must not exceed any window's maxAmount, or no payment could ever fit it",
			)
		}
	}
	return nil
}
