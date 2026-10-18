package requests

import "github.com/google/uuid"

// UpdateMcpConnectionRequest edits a connection; HeaderAuth fully replaces the stored
// set — an omitted header is dropped, and one sent blank keeps its vault value.
type UpdateMcpConnectionRequest struct {
	Id uuid.UUID `uri:"id" json:"-"`
	McpConnectionRequest
	McpHeaderAuthRequest
}

func (r *UpdateMcpConnectionRequest) ValidateAndSanitize() error {
	if err := r.McpConnectionRequest.ValidateAndSanitize(); err != nil {
		return err
	}
	return r.McpHeaderAuthRequest.ValidateAndSanitize()
}
