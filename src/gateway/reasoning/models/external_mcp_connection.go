package models

import "time"

type ExternalMCPConnectionPayload struct {
	Id             string    `json:"id"`
	Name           string    `json:"name"`
	Url            string    `json:"url"`
	HeaderAuthKeys []string  `json:"headerAuthKeys"`
	CreatedAt      time.Time `json:"createdAt"`
	UpdatedAt      time.Time `json:"updatedAt"`
}

type ExternalMCPConnectionListPayload struct {
	Connections []ExternalMCPConnectionPayload `json:"connections"`
}

// ExternalCreateMCPConnectionPayload is the body shape /internal/config/mcp POST accepts.
type ExternalCreateMCPConnectionPayload struct {
	Name       string            `json:"name"`
	Url        string            `json:"url"`
	HeaderAuth map[string]string `json:"headerAuth,omitempty"`
}

// ExternalUpdateMCPConnectionPayload is the body shape /internal/config/mcp/:id PUT accepts.
type ExternalUpdateMCPConnectionPayload struct {
	Name       string            `json:"name"`
	Url        string            `json:"url"`
	HeaderAuth map[string]string `json:"headerAuth,omitempty"`
}
