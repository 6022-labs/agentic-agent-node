package requests

import (
	"strings"

	"github.com/6022-labs/agentic-agent-node/src/agent_payment/models/enums"
	"github.com/6022-labs/agentic-agent-node/src/common/errors"
)

// SavePaymentRuleRequest holds the fields receive and spend rules share; each embeds it and calls it first.
type SavePaymentRuleRequest struct {
	Name   string `json:"name"`
	Method string `json:"method,omitempty"`
}

// ValidateAndSanitize trims the name and defaults the method to x402.
func (a *SavePaymentRuleRequest) ValidateAndSanitize() error {
	a.Name = strings.TrimSpace(a.Name)
	if len(a.Name) == 0 || len(a.Name) > 64 {
		return errors.NewValidationError("name", "must be 1-64 characters")
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
