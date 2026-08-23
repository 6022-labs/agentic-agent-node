package requests

import "github.com/google/uuid"

// DeleteMcpConnectionRequest identifies the MCP connection to delete, decoded from the route params.
type DeleteMcpConnectionRequest struct {
	Id uuid.UUID `uri:"id"`
}
