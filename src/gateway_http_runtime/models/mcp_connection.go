package models

import "time"

type MCPConnectionPayload struct {
	Id             string    `json:"id"`
	Name           string    `json:"name"`
	Url            string    `json:"url"`
	HeaderAuthKeys []string  `json:"headerAuthKeys"`
	CreatedAt      time.Time `json:"createdAt"`
	UpdatedAt      time.Time `json:"updatedAt"`
}

type MCPConnectionListPayload struct {
	Connections []MCPConnectionPayload `json:"connections"`
}

// CreateMCPConnectionPayload is the body shape /internal/config/mcp POST accepts.
type CreateMCPConnectionPayload struct {
	Name       string            `json:"name"`
	Url        string            `json:"url"`
	HeaderAuth map[string]string `json:"headerAuth,omitempty"`
}

// UpdateMCPConnectionPayload is the body shape /internal/config/mcp/:id PUT accepts.
type UpdateMCPConnectionPayload struct {
	Name       string            `json:"name"`
	Url        string            `json:"url"`
	HeaderAuth map[string]string `json:"headerAuth,omitempty"`
}
