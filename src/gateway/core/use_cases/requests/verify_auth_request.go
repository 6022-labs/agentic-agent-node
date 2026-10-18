package requests

import (
	"strings"

	"github.com/6022-labs/agentic-agent-node/src/common/errors"
	"github.com/6022-labs/agentic-agent-node/src/common/helpers"
)

type VerifyAuthRequest struct {
	WalletAddressRequest
	Signature string `json:"signature"`
}

func (r *VerifyAuthRequest) ValidateAndSanitize() error {
	if err := r.WalletAddressRequest.ValidateAndSanitize(); err != nil {
		return err
	}
	r.Signature = strings.TrimSpace(r.Signature)
	if !helpers.IsEthSignature(r.Signature) {
		return errors.NewValidationError("signature", "must be a 0x-prefixed 65-byte hex signature")
	}
	return nil
}
