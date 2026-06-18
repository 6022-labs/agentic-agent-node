package requests

import (
	"strings"

	"github.com/6022-labs/agentic-agent-node/src/common/errors"
)

type SetLlmConfigRequest struct {
	Provider   string         `json:"provider"`
	Model      string         `json:"model"`
	ApiKey     string         `json:"apiKey"`
	BaseUrl    string         `json:"baseUrl,omitempty"`
	Extensions map[string]any `json:"extensions,omitempty"`
}

func (r *SetLlmConfigRequest) ValidateAndSanitize() error {
	if len(strings.TrimSpace(r.Provider)) == 0 {
		return errors.NewValidationError("provider", "cannot be empty")
	}
	if len(strings.TrimSpace(r.Model)) == 0 {
		return errors.NewValidationError("model", "cannot be empty")
	}
	return nil
}
