package requests

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/6022-labs/agentic-agent-node/src/agent_reasoning/utils"
	"github.com/6022-labs/agentic-agent-node/src/common/errors"
)

// MaxMCPConnectionNameLength keeps mcp_<name>_<tool> under the tightest provider function-name limit (~64).
const MaxMCPConnectionNameLength = 15

// A name starts with a letter, then letters/digits/hyphens; no underscores (our separator) or chars providers reject.
var mcpConnectionNamePattern = regexp.MustCompile(`^[a-zA-Z][a-zA-Z0-9-]*$`)

// MCPConnectionRequest is embedded by the requests that create or replace an MCP connection.
type MCPConnectionRequest struct {
	Name                string   `json:"name"`
	Url                 string   `json:"url"`
	AllowedToolPatterns []string `json:"allowedToolPatterns"`
}

func (r *MCPConnectionRequest) ValidateAndSanitize() error {
	r.Name = strings.TrimSpace(r.Name)
	r.Url = strings.TrimSpace(r.Url)

	if len(r.Name) == 0 {
		return errors.NewValidationError("name", "cannot be empty")
	}
	if len(r.Name) > MaxMCPConnectionNameLength {
		return errors.NewValidationError("name", fmt.Sprintf("cannot exceed %d characters", MaxMCPConnectionNameLength))
	}
	if !mcpConnectionNamePattern.MatchString(r.Name) {
		return errors.NewValidationError(
			"name",
			"must start with a letter and contain only letters, digits, and hyphens",
		)
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
		return errors.NewValidationError(
			"allowedToolPatterns",
			"at least one pattern is required (\".*\" allows every tool)",
		)
	}
	for i, pattern := range r.AllowedToolPatterns {
		trimmed := strings.TrimSpace(pattern)
		if len(trimmed) == 0 {
			return errors.NewValidationError("allowedToolPatterns", "patterns cannot be empty")
		}
		if _, err := regexp.Compile(trimmed); err != nil {
			return errors.NewValidationError(
				"allowedToolPatterns",
				"must be valid regexes matched against the tool name",
			)
		}
		r.AllowedToolPatterns[i] = trimmed
	}
	return nil
}
