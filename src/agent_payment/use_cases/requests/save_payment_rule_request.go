package requests

import (
	"encoding/json"
	"math/big"
	"strings"

	"github.com/6022-labs/agentic-agent-node/src/agent_payment/models/enums"
	"github.com/6022-labs/agentic-agent-node/src/common/errors"
)

// SavePaymentRuleRequest is method-neutral: Name/Price/Method are generic;
// Config carries the method-specific settings, validated by that method's
// strategy.
type SavePaymentRuleRequest struct {
	Name   string          `json:"name"`
	Price  string          `json:"price"`
	Method string          `json:"method,omitempty"`
	Config json.RawMessage `json:"config" swaggertype:"object"`
}

// ValidateAndSanitize checks the method-agnostic fields and defaults the method
// to x402. The Config blob is validated by the method's rule validator.
func (a *SavePaymentRuleRequest) ValidateAndSanitize() error {
	a.Name = strings.TrimSpace(a.Name)
	a.Price = strings.TrimSpace(a.Price)

	if len(a.Name) == 0 || len(a.Name) > 64 {
		return errors.NewValidationError("name", "must be 1-64 characters")
	}
	price, ok := new(big.Int).SetString(a.Price, 10)
	if !ok || price.Sign() < 0 {
		return errors.NewValidationError("price", "must be a non-negative integer in the asset's atomic units (0 = free)")
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
