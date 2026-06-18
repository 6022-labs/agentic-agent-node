package x402

// Eip3009Authorization mirrors EIP-3009 TransferWithAuthorization; Nonce is
// 0x-hex 32 bytes, the rest decimal strings.
type Eip3009Authorization struct {
	From        string `json:"from"`
	To          string `json:"to"`
	Value       string `json:"value"`
	ValidAfter  string `json:"validAfter"`
	ValidBefore string `json:"validBefore"`
	Nonce       string `json:"nonce"`
}
