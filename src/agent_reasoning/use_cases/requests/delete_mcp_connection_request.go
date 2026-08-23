package requests

import "github.com/google/uuid"

// DeleteMCPConnectionRequest identifies the MCP connection to delete, decoded from the route params.
type DeleteMCPConnectionRequest struct {
	Id uuid.UUID `uri:"id"`
}
