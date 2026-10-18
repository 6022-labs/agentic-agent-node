package requests

// CreateMCPConnectionRequest registers an MCP connection; every header needs a value since nothing is stored yet.
type CreateMCPConnectionRequest struct {
	MCPConnectionRequest
	MCPHeaderAuthRequest
}

func (r *CreateMCPConnectionRequest) ValidateAndSanitize() error {
	if err := r.MCPConnectionRequest.ValidateAndSanitize(); err != nil {
		return err
	}
	if err := r.MCPHeaderAuthRequest.ValidateAndSanitize(); err != nil {
		return err
	}
	return r.MCPHeaderAuthRequest.ValidateValues()
}
