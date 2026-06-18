package requests

import (
	"strings"

	"github.com/6022-labs/agentic-agent-node/src/common/errors"
)

// ListAgentCollectionsRequest is the chain to list collections for.
type ListAgentCollectionsRequest struct {
	ChainId string `query:"chainId" json:"chainId" binding:"required"`
}

func (r *ListAgentCollectionsRequest) ValidateAndSanitize() error {
	r.ChainId = strings.TrimSpace(r.ChainId)
	if r.ChainId == "" {
		return errors.NewValidationError("chainId", "is required (e.g. 80002)")
	}
	return nil
}
