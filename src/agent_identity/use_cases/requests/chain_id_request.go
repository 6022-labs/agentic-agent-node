package requests

import (
	"strings"

	"github.com/6022-labs/agentic-agent-node/src/common/errors"
)

// ChainIdRequest is embedded by every request addressing one chain by id.
type ChainIdRequest struct {
	ChainId string `query:"chainId" form:"chainId" json:"chainId" binding:"required"`
}

func (r *ChainIdRequest) ValidateAndSanitize() error {
	r.ChainId = strings.TrimSpace(r.ChainId)
	if r.ChainId == "" {
		return errors.NewValidationError("chainId", "is required (e.g. 80002)")
	}
	return nil
}
