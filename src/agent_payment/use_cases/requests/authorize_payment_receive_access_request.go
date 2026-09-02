package requests

import (
	"strings"

	"github.com/6022-labs/agentic-agent-node/src/common/errors"
	"github.com/6022-labs/agentic-agent-node/src/common/helpers"
)

// AuthorizePaymentReceiveAccessRequest carries whichever credential the caller presented; both may be absent.
type AuthorizePaymentReceiveAccessRequest struct {
	// Address is a caller the gateway already proved (JWT); it grants free access only.
	Address          string `json:"address,omitempty"`
	PaymentSignature string `json:"paymentSignature,omitempty"`
}

func (r *AuthorizePaymentReceiveAccessRequest) ValidateAndSanitize() error {
	r.Address = strings.TrimSpace(r.Address)
	r.PaymentSignature = strings.TrimSpace(r.PaymentSignature)

	if len(r.Address) != 0 && !helpers.IsEthAddress(r.Address) {
		return errors.NewValidationError("address", "must be a 0x-prefixed address")
	}
	return nil
}
