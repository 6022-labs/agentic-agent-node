package requests

import (
	"strings"

	"github.com/6022-labs/agentic-agent-node/src/common/errors"
)

// UpdateMCPConnectionRequest replaces all editable fields of an MCP connection
// in one call (full-replacement semantics). HeaderAuth, when present, is the
// new complete set — any previously configured header not in this map is
// dropped on Save.
type UpdateMCPConnectionRequest struct {
	Name       string            `json:"name"`
	Url        string            `json:"url"`
	HeaderAuth map[string]string `json:"headerAuth,omitempty"`
}

func (r *UpdateMCPConnectionRequest) ValidateAndSanitize() error {
	r.Name = strings.TrimSpace(r.Name)
	r.Url = strings.TrimSpace(r.Url)

	if len(r.Name) == 0 {
		return errors.NewValidationError("name", "cannot be empty")
	}
	if len(r.Url) == 0 {
		return errors.NewValidationError("url", "cannot be empty")
	}
	for k, v := range r.HeaderAuth {
		if len(strings.TrimSpace(k)) == 0 {
			return errors.NewValidationError("headerAuth", "header name cannot be empty")
		}
		if len(strings.TrimSpace(v)) == 0 {
			return errors.NewValidationError("headerAuth", "value for "+k+" cannot be empty")
		}
	}
	return nil
}
