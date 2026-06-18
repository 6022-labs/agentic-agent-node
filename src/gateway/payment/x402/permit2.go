package x402

const (
	// Uniswap's canonical Permit2 — same address on every chain via CREATE2.
	Permit2ContractAddress = "0x000000000022D473030F116dDEE9F6B43aC78BA3"
	// Permit2's domain has NO version field — adding one breaks the digest.
	Permit2DomainName = "Permit2"
	// The x402 proxy the signature authorizes to pull the tokens.
	Permit2Spender = "0x402085c248EeA27D92E8b30b2C58ed07f9E20001"
)

type Permit2TokenPermissions struct {
	Token  string `json:"token"`
	Amount string `json:"amount"`
}

// Permit2Witness binds the final recipient and earliest spend time into the signature.
type Permit2Witness struct {
	To         string `json:"to"`
	ValidAfter string `json:"validAfter"`
}

// Permit2Authorization mirrors the signed PermitWitnessTransferFrom; Nonce is a full uint256.
type Permit2Authorization struct {
	Permitted Permit2TokenPermissions `json:"permitted"`
	From      string                  `json:"from"`
	Spender   string                  `json:"spender"`
	Nonce     string                  `json:"nonce"`
	Deadline  string                  `json:"deadline"`
	Witness   Permit2Witness          `json:"witness"`
}
