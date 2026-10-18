package requests

import (
	"strings"

	"github.com/6022-labs/agentic-agent-node/src/common/errors"
	"github.com/google/uuid"
	"github.com/samber/lo"
)

// DiscoverMcpServerToolsRequest describes the server to probe. Every field is optional:
// ConnectionId names a stored connection each missing field falls back to — its url, and
// the vault value of any header sent without one.
type DiscoverMcpServerToolsRequest struct {
	ConnectionId *uuid.UUID `json:"connectionId,omitempty"`
	Url          *string    `json:"url,omitempty"`
	McpHeaderAuthRequest
}

func (r *DiscoverMcpServerToolsRequest) ValidateAndSanitize() error {
	r.Url = lo.EmptyableToPtr(strings.TrimSpace(lo.FromPtr(r.Url)))

	if r.ConnectionId == nil && r.Url == nil {
		return errors.NewValidationError("url", "cannot be empty without a connection id")
	}
	return r.McpHeaderAuthRequest.ValidateAndSanitize()
}
