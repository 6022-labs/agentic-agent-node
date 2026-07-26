package requests

import (
	"strings"

	"github.com/6022-labs/agentic-agent-node/src/common/errors"
	"github.com/google/uuid"
)

// UpdateMcpConnectionRequest edits a connection; HeaderAuth fully replaces the stored
// set — an omitted header is dropped, and one sent blank keeps its vault value.
type UpdateMcpConnectionRequest struct {
	Id                  uuid.UUID         `params:"id" json:"-"`
	Name                string            `json:"name"`
	Url                 string            `json:"url"`
	AllowedToolPatterns []string          `json:"allowedToolPatterns"`
	HeaderAuth          map[string]string `json:"headerAuth,omitempty"`
}

func (r *UpdateMcpConnectionRequest) ValidateAndSanitize() error {
	if len(strings.TrimSpace(r.Name)) == 0 {
		return errors.NewValidationError("name", "cannot be empty")
	}
	if len(strings.TrimSpace(r.Url)) == 0 {
		return errors.NewValidationError("url", "cannot be empty")
	}
	if len(r.AllowedToolPatterns) == 0 {
		return errors.NewValidationError("allowedToolPatterns", "at least one tool must be allowed")
	}
	return validateHeaderAuth(r.HeaderAuth, true)
}
