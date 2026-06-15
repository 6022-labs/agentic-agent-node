package helpers

import (
	"fmt"
	"strconv"
	"strings"
)

// Caip2Eip155Prefix is the CAIP-2 namespace for EVM chains ("eip155:80002").
const Caip2Eip155Prefix = "eip155:"

// ParseCaip2 extracts the chain id from a CAIP-2 eip155 identifier
// ("eip155:80002" → 80002).
func ParseCaip2(network string) (uint64, error) {
	rest, found := strings.CutPrefix(strings.TrimSpace(network), Caip2Eip155Prefix)
	if !found {
		return 0, fmt.Errorf("must be a CAIP-2 eip155 identifier, e.g. eip155:80002")
	}
	chainId, err := strconv.ParseUint(rest, 10, 64)
	if err != nil || chainId == 0 {
		return 0, fmt.Errorf("must be a CAIP-2 eip155 identifier, e.g. eip155:80002")
	}
	return chainId, nil
}

// FormatCaip2 renders a chain id as a CAIP-2 eip155 identifier.
func FormatCaip2(chainId uint64) string {
	return fmt.Sprintf("%s%d", Caip2Eip155Prefix, chainId)
}
