package requests

import (
	"strings"

	"github.com/6022-labs/agentic-agent-node/src/common/errors"
)

// CreateMcpConnectionRequest registers an MCP connection; HeaderAuth carries every
// auth header the server needs, values included since nothing is stored yet.
type CreateMcpConnectionRequest struct {
	Name                string            `json:"name"`
	Url                 string            `json:"url"`
	AllowedToolPatterns []string          `json:"allowedToolPatterns"`
	HeaderAuth          map[string]string `json:"headerAuth,omitempty"`
}

func (r *CreateMcpConnectionRequest) ValidateAndSanitize() error {
	if len(strings.TrimSpace(r.Name)) == 0 {
		return errors.NewValidationError("name", "cannot be empty")
	}
	if len(strings.TrimSpace(r.Url)) == 0 {
		return errors.NewValidationError("url", "cannot be empty")
	}
	if len(r.AllowedToolPatterns) == 0 {
		return errors.NewValidationError("allowedToolPatterns", "at least one tool must be allowed")
	}
	return validateHeaderAuth(r.HeaderAuth, false)
}

// validateHeaderAuth rejects nameless headers, and valueless ones unless the
// caller may fall back to what the vault already holds.
func validateHeaderAuth(headerAuth map[string]string, allowBlankValues bool) error {
	for name, value := range headerAuth {
		if len(strings.TrimSpace(name)) == 0 {
			return errors.NewValidationError("headerAuth", "header name cannot be empty")
		}
		if !allowBlankValues && len(strings.TrimSpace(value)) == 0 {
			return errors.NewValidationError("headerAuth", "value for "+name+" cannot be empty")
		}
	}
	return nil
}
