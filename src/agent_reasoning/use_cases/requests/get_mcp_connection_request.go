package requests

import "github.com/google/uuid"

// GetMCPConnectionRequest identifies the MCP connection to read, decoded from the route params.
type GetMCPConnectionRequest struct {
	Id uuid.UUID `params:"id"`
}
