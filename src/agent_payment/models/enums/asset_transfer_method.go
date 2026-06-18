package enums

// AssetTransferMethod is how an x402 "exact" payment moves the token: the
// token's own EIP-3009 transferWithAuthorization, or Permit2 for any ERC-20.
// It's detected from the asset on-chain at save, never operator input.
type AssetTransferMethod string

const (
	// AssetTransferMethodEip3009 is the token's native EIP-3009 transfer; the
	// buyer signs over the token's own EIP-712 domain.
	AssetTransferMethodEip3009 AssetTransferMethod = "eip3009"
	// AssetTransferMethodPermit2 routes any ERC-20 through Permit2; the buyer
	// signs over Permit2's fixed EIP-712 domain.
	AssetTransferMethodPermit2 AssetTransferMethod = "permit2"
)

func (m AssetTransferMethod) String() string {
	return string(m)
}
