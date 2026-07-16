package requests

import "github.com/google/uuid"

// UpdateMCPConnectionRequest pairs the route id with the replacement connection body.
type UpdateMCPConnectionRequest struct {
	Id uuid.UUID `params:"id" json:"-"`
	SaveMCPConnectionRequest
}
