package requests

import (
	"encoding/json"
	"strings"

	"github.com/6022-labs/agentic-agent-node/src/agent_payment/models/enums"
	"github.com/6022-labs/agentic-agent-node/src/common/errors"
)

// SavePaymentReceiveRuleRequest prices one token: Config names it, validated by the method's strategy.
type SavePaymentReceiveRuleRequest struct {
	Name   string                               `json:"name"`
	Method string                               `json:"method,omitempty"`
	Prices []SavePaymentReceiveRulePriceRequest `json:"prices"`
	Config json.RawMessage                      `json:"config"           swaggertype:"object"`
}

// ValidateAndSanitize checks the method-agnostic fields and defaults the method to x402.
func (a *SavePaymentReceiveRuleRequest) ValidateAndSanitize() error {
	a.Name = strings.TrimSpace(a.Name)
	if len(a.Name) == 0 || len(a.Name) > 64 {
		return errors.NewValidationError("name", "must be 1-64 characters")
	}

	if len(a.Prices) == 0 {
		return errors.NewValidationError("prices", "at least one price is required")
	}
	seen := make(map[string]bool, len(a.Prices))
	for i := range a.Prices {
		if err := a.Prices[i].ValidateAndSanitize(i); err != nil {
			return err
		}
		if seen[a.Prices[i].Price] {
			return errors.NewValidationError("prices", "each price may appear only once")
		}
		seen[a.Prices[i].Price] = true
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

// HasPaidPrice reports whether any price settles on-chain.
func (a *SavePaymentReceiveRuleRequest) HasPaidPrice() bool {
	for i := range a.Prices {
		if a.Prices[i].Price != "0" {
			return true
		}
	}
	return false
}
