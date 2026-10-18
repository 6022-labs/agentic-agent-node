package requests

import (
	"strings"

	"github.com/6022-labs/agentic-agent-node/src/common/errors"
)

// McpHeaderAuthRequest is embedded by every request carrying MCP auth headers.
type McpHeaderAuthRequest struct {
	HeaderAuth map[string]string `json:"headerAuth,omitempty"`
}

func (r *McpHeaderAuthRequest) ValidateAndSanitize() error {
	for name := range r.HeaderAuth {
		if len(strings.TrimSpace(name)) == 0 {
			return errors.NewValidationError("headerAuth", "header name cannot be empty")
		}
	}
	return nil
}

// ValidateValues rejects blank values, for callers with no stored value to fall back to.
func (r *McpHeaderAuthRequest) ValidateValues() error {
	for name, value := range r.HeaderAuth {
		if len(strings.TrimSpace(value)) == 0 {
			return errors.NewValidationError("headerAuth", "value for "+name+" cannot be empty")
		}
	}
	return nil
}
