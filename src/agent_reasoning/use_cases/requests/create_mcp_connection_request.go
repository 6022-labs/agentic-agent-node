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

// CreateMCPConnectionRequest holds an MCP connection's editable fields; a present HeaderAuth fully replaces the stored set, so omitted headers are dropped on save.
type CreateMCPConnectionRequest struct {
	Name                string            `json:"name"`
	Url                 string            `json:"url"`
	AllowedToolPatterns []string          `json:"allowedToolPatterns"`
	HeaderAuth          map[string]string `json:"headerAuth,omitempty"`
}

func (r *CreateMCPConnectionRequest) ValidateAndSanitize() error {
	return r.validateAndSanitize(false)
}

func (r *CreateMCPConnectionRequest) validateAndSanitize(allowEmptyHeaderValues bool) error {
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
	if len(r.AllowedToolPatterns) == 0 {
		return errors.NewValidationError("allowedToolPatterns", "at least one pattern is required (\".*\" allows every tool)")
	}
	for i, pattern := range r.AllowedToolPatterns {
		trimmed := strings.TrimSpace(pattern)
		if len(trimmed) == 0 {
			return errors.NewValidationError("allowedToolPatterns", "patterns cannot be empty")
		}
		if _, err := regexp.Compile(trimmed); err != nil {
			return errors.NewValidationError("allowedToolPatterns", "must be valid regexes matched against the tool name")
		}
		r.AllowedToolPatterns[i] = trimmed
	}
	for k, v := range r.HeaderAuth {
		if len(strings.TrimSpace(k)) == 0 {
			return errors.NewValidationError("headerAuth", "header name cannot be empty")
		}
		if !allowEmptyHeaderValues && len(strings.TrimSpace(v)) == 0 {
			return errors.NewValidationError("headerAuth", "value for "+k+" cannot be empty")
		}
	}
	return nil
}
