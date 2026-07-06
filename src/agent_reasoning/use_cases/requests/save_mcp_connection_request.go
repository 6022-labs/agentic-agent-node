package requests

import (
	"strings"

	"github.com/6022-labs/agentic-agent-node/src/agent_reasoning/utils"
	"github.com/6022-labs/agentic-agent-node/src/common/errors"
)

// SaveMCPConnectionRequest carries an MCP connection's editable fields for
// create and update. HeaderAuth, if present, is the complete new set — headers
// absent from the map are dropped on Save.
type SaveMCPConnectionRequest struct {
	Name       string            `json:"name"`
	Url        string            `json:"url"`
	HeaderAuth map[string]string `json:"headerAuth,omitempty"`
}

func (r *SaveMCPConnectionRequest) ValidateAndSanitize() error {
	r.Name = strings.TrimSpace(r.Name)
	r.Url = strings.TrimSpace(r.Url)

	if len(r.Name) == 0 {
		return errors.NewValidationError("name", "cannot be empty")
	}
	if strings.EqualFold(r.Name, utils.SelfToolNamespace) {
		return errors.NewValidationError("name", "\""+utils.SelfToolNamespace+"\" is reserved for built-in tools")
	}
	if strings.Contains(r.Name, "_") {
		return errors.NewValidationError("name", "cannot contain underscores")
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
