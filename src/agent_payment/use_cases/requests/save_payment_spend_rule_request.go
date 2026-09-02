package requests

import (
	"encoding/json"
	"math/big"
	"strings"

	"github.com/6022-labs/agentic-agent-node/src/agent_payment/models/enums"
	"github.com/6022-labs/agentic-agent-node/src/common/errors"
)

// SavePaymentSpendRuleRequest is method-neutral: Name/MaxAmount/Method are generic.
type SavePaymentSpendRuleRequest struct {
	Name      string          `json:"name"`
	MaxAmount string          `json:"maxAmount"`
	Method    string          `json:"method,omitempty"`
	Config    json.RawMessage `json:"config"           swaggertype:"object"`
}

// ValidateAndSanitize checks the method-agnostic fields and defaults the method to x402.
func (a *SavePaymentSpendRuleRequest) ValidateAndSanitize() error {
	a.Name = strings.TrimSpace(a.Name)
	a.MaxAmount = strings.TrimSpace(a.MaxAmount)

	if len(a.Name) == 0 || len(a.Name) > 64 {
		return errors.NewValidationError("name", "must be 1-64 characters")
	}
	// There is no unlimited sentinel: "0" denies every non-zero quote in scope.
	maxAmount, ok := new(big.Int).SetString(a.MaxAmount, 10)
	if !ok || maxAmount.Sign() < 0 {
		return errors.NewValidationError(
			"maxAmount",
			"must be a non-negative integer in the asset's atomic units (0 = deny)",
		)
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
