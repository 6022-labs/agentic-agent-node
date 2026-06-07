package responses

import (
	"time"

	"github.com/6022-labs/agentic-agent-node/src/agent_reasoning/models"
	reasoning_enums "github.com/6022-labs/agentic-agent-node/src/agent_reasoning/models/enums"
	"github.com/google/uuid"
)

// LLMConnectionResponse is the wire shape returned by /internal/config/llm.
// The api key never appears here — it lives only in Vault.
type LLMConnectionResponse struct {
	Id          uuid.UUID                                `json:"id"`
	Platform    reasoning_enums.LLMConnectionPlatform    `json:"platform"`
	Model       string                                   `json:"model"`
	BaseUrl     *string                                  `json:"baseUrl,omitempty"`
	SetupStatus reasoning_enums.LLMConnectionSetupStatus `json:"setupStatus"`
	CreatedAt   time.Time                                `json:"createdAt"`
	UpdatedAt   time.Time                                `json:"updatedAt"`
}

func ToLLMConnectionResponse(c *models.LLMConnection) *LLMConnectionResponse {
	if c == nil {
		return nil
	}
	return &LLMConnectionResponse{
		Id:          c.Id,
		Platform:    c.Platform,
		Model:       c.Model,
		BaseUrl:     c.BaseUrl,
		SetupStatus: c.SetupStatus,
		CreatedAt:   c.CreatedAt,
		UpdatedAt:   c.UpdatedAt,
	}
}
