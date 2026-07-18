package requests

import (
	"strings"

	"github.com/6022-labs/agentic-agent-node/src/common/errors"
	"github.com/6022-labs/agentic-agent-node/src/common/helpers"
)

// GetAuthChallengeRequest carries the wallet the SIWE message is issued for.
type GetAuthChallengeRequest struct {
	Address string `query:"address"`
	ChainId uint64 `query:"chainId"`
}

func (r *GetAuthChallengeRequest) ValidateAndSanitize() error {
	r.Address = strings.TrimSpace(r.Address)
	if !helpers.IsEthAddress(r.Address) {
		return errors.NewValidationError("address", "must be a 0x-prefixed 20-byte hex address")
	}
	if r.ChainId == 0 {
		return errors.NewValidationError("chainId", "must be a positive chain id")
	}
	return nil
}
