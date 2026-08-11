package requests

import (
	"strings"

	reasoning_enums "github.com/6022-labs/agentic-agent-node/src/agent_reasoning/models/enums"
	"github.com/6022-labs/agentic-agent-node/src/common/errors"
)

type SetupLLMConnectionRequest struct {
	Platform reasoning_enums.LLMConnectionPlatform `json:"platform"`
	Model    string                                `json:"model"`
	BaseUrl  *string                               `json:"baseUrl,omitempty"`
	ApiKey   string                                `json:"apiKey"`
}

func (r *SetupLLMConnectionRequest) ValidateAndSanitize() error {
	r.Model = strings.TrimSpace(r.Model)
	r.ApiKey = strings.TrimSpace(r.ApiKey)
	if r.BaseUrl != nil {
		trimmed := strings.TrimSpace(*r.BaseUrl)
		if len(trimmed) == 0 {
			r.BaseUrl = nil
		} else {
			r.BaseUrl = &trimmed
		}
	}

	if !reasoning_enums.GetAllLLMConnectionPlatforms().Contains(r.Platform) {
		return errors.NewValidationError(
			"platform",
			"must be one of "+reasoning_enums.GetAllLLMConnectionPlatforms().String(),
		)
	}
	if len(r.Model) == 0 {
		return errors.NewValidationError("model", "cannot be empty")
	}
	return nil
}
