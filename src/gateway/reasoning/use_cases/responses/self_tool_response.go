package responses

import (
	"time"

	"github.com/6022-labs/agentic-agent-node/src/gateway/reasoning/models"
)

// SelfToolResponse is the gateway-shaped wire response for self tool endpoints.
type SelfToolResponse struct {
	Id          string    `json:"id"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	Group       string    `json:"group"`
	Enabled     bool      `json:"enabled"`
	CreatedAt   time.Time `json:"createdAt"`
	UpdatedAt   time.Time `json:"updatedAt"`
}

type ListSelfToolsResponse struct {
	Tools []SelfToolResponse `json:"tools"`
}

func ToSelfToolResponse(p *models.ExternalSelfToolPayload) *SelfToolResponse {
	return &SelfToolResponse{
		Id:          p.Id,
		Name:        p.Name,
		Description: p.Description,
		Group:       p.Group,
		Enabled:     p.Enabled,
		CreatedAt:   p.CreatedAt,
		UpdatedAt:   p.UpdatedAt,
	}
}
