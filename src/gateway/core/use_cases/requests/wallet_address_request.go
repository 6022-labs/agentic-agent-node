package requests

import (
	"strings"

	"github.com/6022-labs/agentic-agent-node/src/common/errors"
	"github.com/6022-labs/agentic-agent-node/src/common/helpers"
)

// WalletAddressRequest is embedded by every request naming the wallet it acts for.
type WalletAddressRequest struct {
	Address string `query:"address" json:"address"`
}

func (r *WalletAddressRequest) ValidateAndSanitize() error {
	r.Address = strings.TrimSpace(r.Address)
	if !helpers.IsEthAddress(r.Address) {
		return errors.NewValidationError("address", "must be a 0x-prefixed 20-byte hex address")
	}
	return nil
}
