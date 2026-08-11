package responses

import (
	"time"

	"github.com/6022-labs/agentic-agent-node/src/agent_reasoning/models"
	"github.com/google/uuid"
)

// SelfToolResponse is the wire shape returned by /internal/config/self-tools.
// Description and Group come from the registry, not the DB row.
type SelfToolResponse struct {
	Id          uuid.UUID `json:"id"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	Group       string    `json:"group"`
	Enabled     bool      `json:"enabled"`
	CreatedAt   time.Time `json:"createdAt"`
	UpdatedAt   time.Time `json:"updatedAt"`
}

type SelfToolListResponse struct {
	Tools []SelfToolResponse `json:"tools"`
}

func ToSelfToolResponse(tool *models.SelfTool, description string, group string) *SelfToolResponse {
	if tool == nil {
		return nil
	}
	return &SelfToolResponse{
		Id:          tool.Id,
		Name:        tool.Name,
		Description: description,
		Group:       group,
		Enabled:     tool.Enabled,
		CreatedAt:   tool.CreatedAt,
		UpdatedAt:   tool.UpdatedAt,
	}
}
