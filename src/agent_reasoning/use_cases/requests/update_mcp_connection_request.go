package requests

import "github.com/google/uuid"

// UpdateMCPConnectionRequest pairs the route id with the replacement connection body.
type UpdateMCPConnectionRequest struct {
	Id uuid.UUID `uri:"id" json:"-"`
	CreateMCPConnectionRequest
}

// ValidateAndSanitize allows blank header values: on update they mean "keep the stored value".
func (r *UpdateMCPConnectionRequest) ValidateAndSanitize() error {
	return r.validateAndSanitize(true)
}
