package responses

import (
	"time"

	"github.com/6022-labs/agentic-agent-node/src/gateway/reasoning/models"
)

// McpConnectionResponse is the gateway-shaped wire response for MCP connection
// endpoints. Header auth values stay in Vault and never appear here — only the
// configured key names.
type McpConnectionResponse struct {
	Id                  string    `json:"id"`
	Name                string    `json:"name"`
	Url                 string    `json:"url"`
	AllowedToolPatterns []string  `json:"allowedToolPatterns"`
	HeaderAuthKeys      []string  `json:"headerAuthKeys"`
	CreatedAt           time.Time `json:"createdAt"`
	UpdatedAt           time.Time `json:"updatedAt"`
}

type ListMcpConnectionsResponse struct {
	Connections []McpConnectionResponse `json:"connections"`
}

func ToMcpConnectionResponse(p *models.ExternalMCPConnectionPayload) *McpConnectionResponse {
	return &McpConnectionResponse{
		Id:                  p.Id,
		Name:                p.Name,
		Url:                 p.Url,
		AllowedToolPatterns: p.AllowedToolPatterns,
		HeaderAuthKeys:      p.HeaderAuthKeys,
		CreatedAt:           p.CreatedAt,
		UpdatedAt:           p.UpdatedAt,
	}
}
