package requests

import (
	"regexp"
	"strings"

	"github.com/6022-labs/agentic-agent-node/src/common/errors"
)

var ethAddressPattern = regexp.MustCompile(`^0x[0-9a-fA-F]{40}$`)

type AuthorizeRequest struct {
	Address string `json:"address"`
}

func (r *AuthorizeRequest) ValidateAndSanitize() error {
	r.Address = strings.TrimSpace(r.Address)
	if !ethAddressPattern.MatchString(r.Address) {
		return errors.NewValidationError("address", "must be a 0x-prefixed 20-byte hex address")
	}
	return nil
}
