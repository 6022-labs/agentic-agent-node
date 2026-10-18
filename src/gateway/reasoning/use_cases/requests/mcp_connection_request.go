package requests

import (
	"strings"

	"github.com/6022-labs/agentic-agent-node/src/common/errors"
)

// McpConnectionRequest is embedded by the requests that create or replace an MCP connection.
type McpConnectionRequest struct {
	Name                string   `json:"name"`
	Url                 string   `json:"url"`
	AllowedToolPatterns []string `json:"allowedToolPatterns"`
}

func (r *McpConnectionRequest) ValidateAndSanitize() error {
	if len(strings.TrimSpace(r.Name)) == 0 {
		return errors.NewValidationError("name", "cannot be empty")
	}
	if len(strings.TrimSpace(r.Url)) == 0 {
		return errors.NewValidationError("url", "cannot be empty")
	}
	if len(r.AllowedToolPatterns) == 0 {
		return errors.NewValidationError("allowedToolPatterns", "at least one tool must be allowed")
	}
	return nil
}
