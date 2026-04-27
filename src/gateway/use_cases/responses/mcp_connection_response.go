package responses

import "time"

// McpConnectionResponse is the gateway-shaped wire response for MCP connection
// endpoints. Header auth values stay in Vault and never appear here — only the
// configured key names.
type McpConnectionResponse struct {
	Id             string    `json:"id"`
	Name           string    `json:"name"`
	Url            string    `json:"url"`
	HeaderAuthKeys []string  `json:"headerAuthKeys"`
	CreatedAt      time.Time `json:"createdAt"`
	UpdatedAt      time.Time `json:"updatedAt"`
}

type ListMcpConnectionsResponse struct {
	Connections []McpConnectionResponse `json:"connections"`
}
