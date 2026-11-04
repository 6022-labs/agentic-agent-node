package enums

import (
	"fmt"
	"strings"
)

// AssetTransferMethod is how an x402 payment moves the token: read on-chain for receive, an owner-set filter for spend.
type AssetTransferMethod string

const (
	// AssetTransferMethodEip3009 signs over the token's own EIP-712 domain.
	AssetTransferMethodEip3009 AssetTransferMethod = "eip3009"
	// AssetTransferMethodPermit2 signs over Permit2's fixed EIP-712 domain.
	AssetTransferMethodPermit2 AssetTransferMethod = "permit2"
)

func (m AssetTransferMethod) String() string {
	return string(m)
}

func GetAllAssetTransferMethods() []AssetTransferMethod {
	return []AssetTransferMethod{AssetTransferMethodEip3009, AssetTransferMethodPermit2}
}

func (m *AssetTransferMethod) Set(value string) error {
	normalized := AssetTransferMethod(strings.ToLower(strings.TrimSpace(value)))
	for _, candidate := range GetAllAssetTransferMethods() {
		if normalized == candidate {
			*m = candidate
			return nil
		}
	}
	return fmt.Errorf("invalid AssetTransferMethod: %s", value)
}
