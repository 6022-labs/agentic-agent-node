package requests

import (
	"fmt"
	"math/big"

	"github.com/6022-labs/agentic-agent-node/src/agent_payment/helpers"
	"github.com/6022-labs/agentic-agent-node/src/common/errors"
)

const MaxPaymentSpendWindowSeconds = 10 * 365 * 24 * 60 * 60

const MaxPaymentSpendRuleWindows = 10

// SavePaymentSpendRuleWindowRequest is one recurring budget of a spend rule.
type SavePaymentSpendRuleWindowRequest struct {
	WindowSeconds int    `json:"windowSeconds"`
	MaxAmount     string `json:"maxAmount"`
}

func (w *SavePaymentSpendRuleWindowRequest) ValidateAndSanitize(index int) (*big.Int, error) {
	// The cap keeps the period's time.Duration from overflowing into a window that never accumulates.
	if w.WindowSeconds <= 0 || w.WindowSeconds > MaxPaymentSpendWindowSeconds {
		return nil, errors.NewValidationError(
			fmt.Sprintf("windows[%d].windowSeconds", index),
			fmt.Sprintf(
				"must be a positive number of seconds, at most %d years",
				MaxPaymentSpendWindowSeconds/(365*24*60*60),
			),
		)
	}

	maxAmount, err := helpers.ParseSpendAmount(fmt.Sprintf("windows[%d].maxAmount", index), w.MaxAmount)
	if err != nil {
		return nil, err
	}
	w.MaxAmount = maxAmount.Dec()
	return maxAmount.ToBig(), nil
}
