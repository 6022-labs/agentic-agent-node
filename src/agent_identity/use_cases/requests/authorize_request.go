package requests

import (
	"strings"

	"github.com/6022-labs/agentic-agent-node/src/common/errors"
	"github.com/6022-labs/agentic-agent-node/src/common/helpers"
)

type AuthorizeRequest struct {
	Address string `json:"address"`
}

func (r *AuthorizeRequest) ValidateAndSanitize() error {
	r.Address = strings.TrimSpace(r.Address)
	if !helpers.IsEthAddress(r.Address) {
		return errors.NewValidationError("address", "must be a 0x-prefixed 20-byte hex address")
	}
	return nil
}
