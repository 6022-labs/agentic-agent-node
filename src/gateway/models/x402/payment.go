// Package x402 holds the wire types of the x402 v2 payment protocol
// (github.com/x402-foundation/x402, specs/x402-specification-v2.md).
package x402

const (
	Version     = 2
	SchemeExact = "exact"
)

// PaymentRequirements is one way the resource can be paid for. Amount is in
// the asset's atomic units; Network is a CAIP-2 identifier. Extra carries the
// asset's EIP-712 domain ({"name", "version"}) so clients can sign EIP-3009
// transfers.
type PaymentRequirements struct {
	Scheme            string         `json:"scheme"`
	Network           string         `json:"network"`
	Amount            string         `json:"amount"`
	Asset             string         `json:"asset"`
	PayTo             string         `json:"payTo"`
	MaxTimeoutSeconds int            `json:"maxTimeoutSeconds"`
	Extra             map[string]any `json:"extra,omitempty"`
}

type ResourceInfo struct {
	Url         string `json:"url"`
	Description string `json:"description,omitempty"`
	MimeType    string `json:"mimeType,omitempty"`
}

// PaymentRequired is the 402 response body, also carried base64-encoded in
// the PAYMENT-REQUIRED header.
type PaymentRequired struct {
	X402Version int                   `json:"x402Version"`
	Resource    ResourceInfo          `json:"resource"`
	Accepts     []PaymentRequirements `json:"accepts"`
	Error       string                `json:"error,omitempty"`
}

// ExactEvmAuthorization mirrors EIP-3009 TransferWithAuthorization fields.
// Value/ValidAfter/ValidBefore are decimal strings; Nonce is 0x-hex 32 bytes.
type ExactEvmAuthorization struct {
	From        string `json:"from"`
	To          string `json:"to"`
	Value       string `json:"value"`
	ValidAfter  string `json:"validAfter"`
	ValidBefore string `json:"validBefore"`
	Nonce       string `json:"nonce"`
}

type ExactEvmPayload struct {
	Signature     string                `json:"signature"`
	Authorization ExactEvmAuthorization `json:"authorization"`
}

// PaymentPayload is the decoded PAYMENT-SIGNATURE header: the client's
// scheme/network selection plus the signed authorization.
type PaymentPayload struct {
	X402Version int             `json:"x402Version"`
	Scheme      string          `json:"scheme"`
	Network     string          `json:"network"`
	Payload     ExactEvmPayload `json:"payload"`
}

// VerifyResponse is the facilitator's POST /verify result.
type VerifyResponse struct {
	IsValid       bool   `json:"isValid"`
	InvalidReason string `json:"invalidReason,omitempty"`
	Payer         string `json:"payer,omitempty"`
}

// SettleResponse is the facilitator's POST /settle result, returned to the
// client base64-encoded in the PAYMENT-RESPONSE header.
type SettleResponse struct {
	Success     bool   `json:"success"`
	ErrorReason string `json:"errorReason,omitempty"`
	Transaction string `json:"transaction,omitempty"`
	Network     string `json:"network,omitempty"`
	Payer       string `json:"payer,omitempty"`
}
