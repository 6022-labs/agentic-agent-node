package requests

import (
	"math/big"
	"strings"

	"github.com/6022-labs/agentic-agent-node/src/agent_payment/models/enums"
	"github.com/6022-labs/agentic-agent-node/src/common/errors"
	"github.com/6022-labs/agentic-agent-node/src/common/helpers"
)

// Extra keys an x402 quote carries the asset's EIP-712 domain under.
const (
	extraAssetTransferMethodKey = "assetTransferMethod"
	extraNameKey                = "name"
	extraVersionKey             = "version"
)

// SignX402PaymentRequest is one `accepts` entry from a 402; the method and EIP-712 domain fall back to Extra.
type SignX402PaymentRequest struct {
	Network             string         `json:"network"`
	Asset               string         `json:"asset"`
	PayTo               string         `json:"payTo"`
	Amount              string         `json:"amount"`
	AssetTransferMethod string         `json:"assetTransferMethod,omitempty"`
	Eip712Name          string         `json:"eip712Name,omitempty"`
	Eip712Version       string         `json:"eip712Version,omitempty"`
	Extra               map[string]any `json:"extra,omitempty"`

	parsedAssetTransferMethod enums.AssetTransferMethod
}

// ParsedAssetTransferMethod is set by ValidateAndSanitize.
func (r *SignX402PaymentRequest) ParsedAssetTransferMethod() enums.AssetTransferMethod {
	return r.parsedAssetTransferMethod
}

func (r *SignX402PaymentRequest) ValidateAndSanitize() error {
	r.Network = strings.TrimSpace(r.Network)
	r.Asset = strings.TrimSpace(r.Asset)
	r.PayTo = strings.TrimSpace(r.PayTo)
	r.Amount = strings.TrimSpace(r.Amount)

	if _, err := helpers.ParseCaip2(r.Network); err != nil {
		return errors.NewValidationError("network", err.Error())
	}
	if !helpers.IsEthAddress(r.Asset) {
		return errors.NewValidationError("asset", "must be a 0x-prefixed address")
	}
	if !helpers.IsEthAddress(r.PayTo) {
		return errors.NewValidationError("payTo", "must be a 0x-prefixed address")
	}
	if amount, ok := new(big.Int).SetString(r.Amount, 10); !ok || amount.Sign() < 0 {
		return errors.NewValidationError("amount", "must be a non-negative atomic amount")
	}

	r.applyExtra()

	r.AssetTransferMethod = r.firstNonEmpty(r.AssetTransferMethod, enums.AssetTransferMethodEip3009.String())
	if err := r.parsedAssetTransferMethod.Set(r.AssetTransferMethod); err != nil {
		return errors.NewValidationError("assetTransferMethod", "must be one of eip3009, permit2")
	}
	// A guessed eip3009 domain signs a digest that recovers to another address; permit2 has its own fixed domain.
	if r.parsedAssetTransferMethod == enums.AssetTransferMethodEip3009 &&
		(len(r.Eip712Name) == 0 || len(r.Eip712Version) == 0) {
		return errors.NewValidationError(
			"eip712Name",
			"eip712Name and eip712Version are required for eip3009 (a quote carries them under extra)",
		)
	}
	return nil
}

// applyExtra fills the fields a raw 402 quote only carries under `extra`.
func (r *SignX402PaymentRequest) applyExtra() {
	r.AssetTransferMethod = r.firstNonEmpty(r.AssetTransferMethod, r.extraString(extraAssetTransferMethodKey))
	r.Eip712Name = r.firstNonEmpty(r.Eip712Name, r.extraString(extraNameKey))
	r.Eip712Version = r.firstNonEmpty(r.Eip712Version, r.extraString(extraVersionKey))
}

// extraString reads a string out of the quote's extra map; a missing or non-string value reads as empty.
func (r *SignX402PaymentRequest) extraString(key string) string {
	value, found := r.Extra[key]
	if !found {
		return ""
	}
	text, ok := value.(string)
	if !ok {
		return ""
	}
	return strings.TrimSpace(text)
}

func (r *SignX402PaymentRequest) firstNonEmpty(values ...string) string {
	for _, value := range values {
		if trimmed := strings.TrimSpace(value); len(trimmed) != 0 {
			return trimmed
		}
	}
	return ""
}
