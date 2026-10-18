package requests

import "github.com/6022-labs/agentic-agent-node/src/common/errors"

// GetAuthChallengeRequest carries the wallet the SIWE message is issued for.
type GetAuthChallengeRequest struct {
	WalletAddressRequest
	ChainId uint64 `query:"chainId"`
}

func (r *GetAuthChallengeRequest) ValidateAndSanitize() error {
	if err := r.WalletAddressRequest.ValidateAndSanitize(); err != nil {
		return err
	}
	if r.ChainId == 0 {
		return errors.NewValidationError("chainId", "must be a positive chain id")
	}
	return nil
}
