package requests

import (
	"encoding/json"
	"math/big"
	"strings"

	"github.com/6022-labs/agentic-agent-node/src/agent_payment/models/enums"
	"github.com/6022-labs/agentic-agent-node/src/common/errors"
)

// SavePaymentSpendRuleRequest is method-neutral: the bounds, name and method are generic.
type SavePaymentSpendRuleRequest struct {
	Name                string          `json:"name"`
	MaxAmountPerPayment string          `json:"maxAmountPerPayment"`
	WindowSeconds       int             `json:"windowSeconds,omitempty"`
	MaxAmountPerWindow  string          `json:"maxAmountPerWindow,omitempty"`
	Method              string          `json:"method,omitempty"`
	Config              json.RawMessage `json:"config"                       swaggertype:"object"`
}

// ValidateAndSanitize checks the method-agnostic fields and defaults the method to x402.
func (a *SavePaymentSpendRuleRequest) ValidateAndSanitize() error {
	a.Name = strings.TrimSpace(a.Name)
	a.MaxAmountPerPayment = strings.TrimSpace(a.MaxAmountPerPayment)

	if len(a.Name) == 0 || len(a.Name) > 64 {
		return errors.NewValidationError("name", "must be 1-64 characters")
	}
	// There is no unlimited sentinel: "0" denies every non-zero quote in scope.
	maxAmountPerPayment, ok := new(big.Int).SetString(a.MaxAmountPerPayment, 10)
	if !ok || maxAmountPerPayment.Sign() < 0 {
		return errors.NewValidationError(
			"maxAmountPerPayment",
			"must be a non-negative integer in the asset's atomic units (0 = deny)",
		)
	}

	if err := a.validateWindow(maxAmountPerPayment); err != nil {
		return err
	}

	var method enums.PaymentMethod
	if len(strings.TrimSpace(a.Method)) == 0 {
		method = enums.PaymentMethodX402
	} else if err := method.Set(a.Method); err != nil {
		return errors.NewValidationError("method", "must be one of: "+enums.GetAllPaymentMethods().String())
	}
	a.Method = method.String()

	return nil
}

// validateWindow requires a period and a budget together, with the per-payment cap within the budget.
func (a *SavePaymentSpendRuleRequest) validateWindow(maxAmountPerPayment *big.Int) error {
	a.MaxAmountPerWindow = strings.TrimSpace(a.MaxAmountPerWindow)

	if a.WindowSeconds < 0 {
		return errors.NewValidationError("windowSeconds", "must not be negative")
	}
	if a.WindowSeconds == 0 {
		if len(a.MaxAmountPerWindow) != 0 {
			return errors.NewValidationError(
				"maxAmountPerWindow",
				"needs a windowSeconds period to be spent over",
			)
		}
		a.MaxAmountPerWindow = "0"
		return nil
	}

	maxAmountPerWindow, ok := new(big.Int).SetString(a.MaxAmountPerWindow, 10)
	if !ok || maxAmountPerWindow.Sign() < 0 {
		return errors.NewValidationError(
			"maxAmountPerWindow",
			"must be a non-negative integer in the asset's atomic units (0 = deny)",
		)
	}
	if maxAmountPerPayment.Cmp(maxAmountPerWindow) > 0 {
		return errors.NewValidationError(
			"maxAmountPerPayment",
			"must not exceed maxAmountPerWindow, or no payment could ever fit the period",
		)
	}
	return nil
}
