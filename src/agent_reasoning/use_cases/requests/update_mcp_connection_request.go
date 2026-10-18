package requests

import "github.com/google/uuid"

// UpdateMCPConnectionRequest allows blank header values: on update they mean "keep the stored value".
type UpdateMCPConnectionRequest struct {
	Id uuid.UUID `uri:"id" json:"-"`
	MCPConnectionRequest
	MCPHeaderAuthRequest
}

func (r *UpdateMCPConnectionRequest) ValidateAndSanitize() error {
	if err := r.MCPConnectionRequest.ValidateAndSanitize(); err != nil {
		return err
	}
	return r.MCPHeaderAuthRequest.ValidateAndSanitize()
}
