package models

// ExternalMcpToolCallResult is the CallToolResult payload; an absent isError means the tool succeeded.
type ExternalMcpToolCallResult struct {
	IsError bool `json:"isError"`
}
