package helpers

import (
	"regexp"
	"strings"
)

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
