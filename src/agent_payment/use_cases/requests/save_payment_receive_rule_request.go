package requests

import (
	"encoding/json"

	"github.com/6022-labs/agentic-agent-node/src/common/errors"
)

// SavePaymentReceiveRuleRequest prices one token: Config names it, validated by the method's strategy.
type SavePaymentReceiveRuleRequest struct {
	SavePaymentRuleRequest
	Prices []SavePaymentReceiveRulePriceRequest `json:"prices"`
	Config json.RawMessage                      `json:"config" swaggertype:"object"`
}

func (a *SavePaymentReceiveRuleRequest) ValidateAndSanitize() error {
	if err := a.SavePaymentRuleRequest.ValidateAndSanitize(); err != nil {
		return err
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
