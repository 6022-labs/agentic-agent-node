package requests

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/6022-labs/agentic-agent-node/src/agent_reasoning/utils"
	"github.com/6022-labs/agentic-agent-node/src/common/errors"
)

// MaxMCPConnectionNameLength keeps mcp_<name>_<tool> under the tightest provider function-name limit (~64), leaving room for the server-supplied tool name.
const MaxMCPConnectionNameLength = 15

// A name must start with a letter, then letters/digits/hyphens only — no underscores (our separator) or characters the provider function-name regex would reject.
var mcpConnectionNamePattern = regexp.MustCompile(`^[a-zA-Z][a-zA-Z0-9-]*$`)

// SaveMCPConnectionRequest holds an MCP connection's editable fields; a present HeaderAuth fully replaces the stored set, so omitted headers are dropped on save.
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
	if len(r.Name) > MaxMCPConnectionNameLength {
		return errors.NewValidationError("name", fmt.Sprintf("cannot exceed %d characters", MaxMCPConnectionNameLength))
	}
	if !mcpConnectionNamePattern.MatchString(r.Name) {
		return errors.NewValidationError("name", "must start with a letter and contain only letters, digits, and hyphens")
	}
	if strings.HasPrefix(strings.ToLower(r.Name), utils.McpToolNamespace) {
		return errors.NewValidationError("name", "cannot start with \""+utils.McpToolNamespace+"\"")
	}
	if strings.EqualFold(r.Name, utils.SelfToolNamespace) {
		return errors.NewValidationError("name", "\""+utils.SelfToolNamespace+"\" is reserved for built-in tools")
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
