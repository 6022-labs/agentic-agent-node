package helpers

import (
	"regexp"
	"strings"

	"github.com/6022-labs/agentic-agent-node/src/common/errors"
	"github.com/ethereum/go-ethereum/common"
)

// everyonePattern covers every address, so any pattern beside it is noise.
const everyonePattern = ".*"

// MatchesAddressPatterns full-matches the lowercased address; patterns run as written, so uppercase hex never matches.
func MatchesAddressPatterns(patterns []string, address string) bool {
	lowered := strings.ToLower(strings.TrimSpace(address))
	for _, pattern := range patterns {
		compiled, err := regexp.Compile("^(?:" + pattern + ")$")
		if err != nil {
			continue
		}
		if compiled.MatchString(lowered) {
			return true
		}
	}
	return false
}

// SanitizeAddressPatterns trims, validates, lowercases plain addresses and dedupes; errors name field.
func SanitizeAddressPatterns(field string, patterns []string) ([]string, error) {
	sanitized := make([]string, 0, len(patterns))
	seen := make(map[string]bool, len(patterns))
	for _, pattern := range patterns {
		trimmed := strings.TrimSpace(pattern)
		if len(trimmed) == 0 {
			return nil, errors.NewValidationError(field, "patterns cannot be empty")
		}
		if _, err := regexp.Compile(trimmed); err != nil {
			return nil, errors.NewValidationError(field, "must be addresses or valid regexes")
		}
		// A pasted checksummed address would compile and then never match the lowercased address.
		if common.IsHexAddress(trimmed) {
			trimmed = strings.ToLower(trimmed)
		}
		if seen[trimmed] {
			continue
		}
		seen[trimmed] = true
		sanitized = append(sanitized, trimmed)
	}
	if seen[everyonePattern] {
		return []string{everyonePattern}, nil
	}
	return sanitized, nil
}
