package requests

import (
	"regexp"
	"strings"

	"github.com/6022-labs/agentic-agent-node/src/common/errors"
	"github.com/ethereum/go-ethereum/common"
)

// UpdatePaymentSpendPolicyRequest replaces who the agent may pay; an empty list pays nobody.
type UpdatePaymentSpendPolicyRequest struct {
	PayToPatterns []string `json:"payToPatterns"`
}

func (r *UpdatePaymentSpendPolicyRequest) ValidateAndSanitize() error {
	sanitized := make([]string, 0, len(r.PayToPatterns))
	seen := make(map[string]bool, len(r.PayToPatterns))
	for _, pattern := range r.PayToPatterns {
		trimmed := strings.TrimSpace(pattern)
		if len(trimmed) == 0 {
			return errors.NewValidationError("payToPatterns", "patterns cannot be empty")
		}
		if _, err := regexp.Compile(trimmed); err != nil {
			return errors.NewValidationError(
				"payToPatterns",
				"must be valid regexes matched against the payee address",
			)
		}
		// A pasted checksummed address would compile and then never match the lowercased payee.
		if common.IsHexAddress(trimmed) {
			trimmed = strings.ToLower(trimmed)
		}
		if seen[trimmed] {
			continue
		}
		seen[trimmed] = true
		sanitized = append(sanitized, trimmed)
	}
	r.PayToPatterns = sanitized
	return nil
}
