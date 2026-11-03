package helpers

import (
	"strings"

	"github.com/6022-labs/agentic-agent-node/src/common/errors"
	"github.com/holiman/uint256"
)

// ParseSpendAmount parses a spend limit in atomic units, refusing it on field when it is not a uint256.
func ParseSpendAmount(field string, value string) (*uint256.Int, error) {
	amount, err := uint256.FromDecimal(strings.TrimSpace(value))
	if err != nil {
		return nil, errors.NewValidationError(
			field,
			"must be a non-negative integer in the asset's atomic units (0 = deny)",
		)
	}
	return amount, nil
}
