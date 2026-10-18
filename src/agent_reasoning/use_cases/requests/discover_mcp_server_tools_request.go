package requests

import (
	"strings"

	"github.com/6022-labs/agentic-agent-node/src/common/errors"
	"github.com/google/uuid"
	"github.com/samber/lo"
)

// DiscoverMCPServerToolsRequest describes the server to probe. Every field is
// optional: ConnectionId names a stored connection each missing field falls back to.
type DiscoverMCPServerToolsRequest struct {
	ConnectionId *uuid.UUID `json:"connectionId,omitempty"`
	Url          *string    `json:"url,omitempty"`
	MCPHeaderAuthRequest
}

func (r *DiscoverMCPServerToolsRequest) ValidateAndSanitize() error {
	r.Url = lo.EmptyableToPtr(strings.TrimSpace(lo.FromPtr(r.Url)))

	if r.ConnectionId == nil && r.Url == nil {
		return errors.NewValidationError("url", "cannot be empty without a connection id")
	}
	return r.MCPHeaderAuthRequest.ValidateAndSanitize()
}
