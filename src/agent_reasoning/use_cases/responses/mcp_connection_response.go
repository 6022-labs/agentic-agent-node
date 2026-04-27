package responses

import (
	"time"

	"github.com/6022-labs/agentic-agent-node/src/agent_reasoning/models"
	"github.com/google/uuid"
)

// MCPConnectionResponse is the wire shape returned by /internal/config/mcp.
// HeaderAuthKeys lists the header *names* the connection authenticates with;
// the values stay in Vault and never appear in any API response.
type MCPConnectionResponse struct {
	Id             uuid.UUID `json:"id"`
	Name           string    `json:"name"`
	Url            string    `json:"url"`
	HeaderAuthKeys []string  `json:"headerAuthKeys"`
	CreatedAt      time.Time `json:"createdAt"`
	UpdatedAt      time.Time `json:"updatedAt"`
}

type MCPConnectionListResponse struct {
	Connections []MCPConnectionResponse `json:"connections"`
}

func ToMCPConnectionResponse(c *models.MCPConnection) *MCPConnectionResponse {
	if c == nil {
		return nil
	}
	keys := make([]string, 0, len(c.HeaderAuthKeys))
	for _, k := range c.HeaderAuthKeys {
		keys = append(keys, string(k))
	}
	return &MCPConnectionResponse{
		Id:             c.Id,
		Name:           c.Name,
		Url:            c.Url,
		HeaderAuthKeys: keys,
		CreatedAt:      c.CreatedAt,
		UpdatedAt:      c.UpdatedAt,
	}
}

func ToMCPConnectionListResponse(connections []models.MCPConnection) *MCPConnectionListResponse {
	out := make([]MCPConnectionResponse, 0, len(connections))
	for i := range connections {
		out = append(out, *ToMCPConnectionResponse(&connections[i]))
	}
	return &MCPConnectionListResponse{Connections: out}
}
