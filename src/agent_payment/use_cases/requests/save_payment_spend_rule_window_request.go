package requests

import (
	"fmt"
	"math/big"
	"strings"

	"github.com/6022-labs/agentic-agent-node/src/common/errors"
	"github.com/6022-labs/agentic-agent-node/src/common/helpers"
)

const MaxPaymentSpendWindowSeconds = 10 * 365 * 24 * 60 * 60

// SavePaymentSpendRuleWindowRequest is one recurring budget of a spend rule.
type SavePaymentSpendRuleWindowRequest struct {
	WindowSeconds int    `json:"windowSeconds"`
	MaxAmount     string `json:"maxAmount"`
}

// ValidateAndSanitize checks the window at index in its rule and returns its parsed budget.
func (w *SavePaymentSpendRuleWindowRequest) ValidateAndSanitize(index int) (*big.Int, error) {
	// The cap keeps the period's time.Duration from overflowing into a window that never accumulates.
	if w.WindowSeconds <= 0 || w.WindowSeconds > MaxPaymentSpendWindowSeconds {
		return nil, errors.NewValidationError(
			fmt.Sprintf("windows[%d].windowSeconds", index),
			"must be a positive number of seconds, at most 10 years",
		)
	}

	w.MaxAmount = strings.TrimSpace(w.MaxAmount)
	maxAmount, ok := helpers.ParseUint256(w.MaxAmount)
	if !ok {
		return nil, errors.NewValidationError(
			fmt.Sprintf("windows[%d].maxAmount", index),
			"must be a non-negative integer in the asset's atomic units (0 = deny)",
		)
	}
	return maxAmount, nil
}
