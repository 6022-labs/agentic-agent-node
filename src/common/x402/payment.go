// Package x402 holds the wire types of the x402 v2 payment protocol
// (github.com/x402-foundation/x402, specs/x402-specification-v2.md).
package x402

const (
	Version     = 2
	SchemeExact = "exact"

	// Mirror the runtime's enums.AssetTransferMethod (gateway can't import it).
	AssetTransferMethodEip3009 = "eip3009"
	AssetTransferMethodPermit2 = "permit2"
)

// PaymentRequirements is one accepted way to pay. Extra carries the asset's
// EIP-712 domain ({name, version}) for eip3009 signing; omitted for permit2.
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

// PaymentRequired is the 402 body, also base64-encoded in the PAYMENT-REQUIRED header.
type PaymentRequired struct {
	X402Version int                   `json:"x402Version"`
	Resource    ResourceInfo          `json:"resource"`
	Accepts     []PaymentRequirements `json:"accepts"`
	Error       string                `json:"error,omitempty"`
}

// ExactEvmPayload carries exactly one signed authorization, per the asset's
// transfer method: eip3009 fills Authorization, permit2 fills
// Permit2Authorization. The unused one stays nil and is omitted from the wire.
type ExactEvmPayload struct {
	Signature            string                `json:"signature"`
	Eip3009Authorization *Eip3009Authorization `json:"authorization,omitempty"`
	Permit2Authorization *Permit2Authorization `json:"permit2Authorization,omitempty"`
}

// PaymentPayload is the decoded PAYMENT-SIGNATURE header.
type PaymentPayload struct {
	X402Version int             `json:"x402Version"`
	Scheme      string          `json:"scheme"`
	Network     string          `json:"network"`
	Payload     ExactEvmPayload `json:"payload"`
}

// RecipientAddress is the payee, method-neutral: EIP-3009 `to` or Permit2 witness `to`.
func (p *PaymentPayload) RecipientAddress() string {
	if p.Payload.Permit2Authorization != nil {
		return p.Payload.Permit2Authorization.Witness.To
	}
	if p.Payload.Eip3009Authorization != nil {
		return p.Payload.Eip3009Authorization.To
	}
	return ""
}

// AmountFor reads only the field the given method signs, never the other one.
func (p *PaymentPayload) AmountFor(assetTransferMethod string) string {
	switch assetTransferMethod {
	case AssetTransferMethodPermit2:
		if p.Payload.Permit2Authorization != nil {
			return p.Payload.Permit2Authorization.Permitted.Amount
		}
	case AssetTransferMethodEip3009:
		if p.Payload.Eip3009Authorization != nil {
			return p.Payload.Eip3009Authorization.Value
		}
	}
	return ""
}

// VerifyResponse is the facilitator's POST /verify result.
type VerifyResponse struct {
	IsValid       bool   `json:"isValid"`
	InvalidReason string `json:"invalidReason,omitempty"`
	Payer         string `json:"payer,omitempty"`
}

// SettleResponse is the facilitator's POST /settle result, base64-encoded back
// to the client in the PAYMENT-RESPONSE header.
type SettleResponse struct {
	Success     bool   `json:"success"`
	ErrorReason string `json:"errorReason,omitempty"`
	Transaction string `json:"transaction,omitempty"`
	Network     string `json:"network,omitempty"`
	Payer       string `json:"payer,omitempty"`
}
