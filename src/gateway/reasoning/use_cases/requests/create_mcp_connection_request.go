package requests

import (
	"strings"

	"github.com/6022-labs/agentic-agent-node/src/common/errors"
)

type CreateMcpConnectionRequest struct {
	Name            string `json:"name"`
	Url             string `json:"url"`
	AuthHeaderName  string `json:"authHeaderName,omitempty"`
	AuthHeaderValue string `json:"authHeaderValue,omitempty"`
}

func (r *CreateMcpConnectionRequest) ValidateAndSanitize() error {
	if len(strings.TrimSpace(r.Name)) == 0 {
		return errors.NewValidationError("name", "cannot be empty")
	}
	if len(strings.TrimSpace(r.Url)) == 0 {
		return errors.NewValidationError("url", "cannot be empty")
	}
	if len(strings.TrimSpace(r.AuthHeaderName)) > 0 && len(strings.TrimSpace(r.AuthHeaderValue)) == 0 {
		return errors.NewValidationError("authHeaderValue", "cannot be empty when authHeaderName is set")
	}
	return nil
}
