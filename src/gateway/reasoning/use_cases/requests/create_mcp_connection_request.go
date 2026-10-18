package requests

// CreateMcpConnectionRequest registers an MCP connection; every header needs a value since nothing is stored yet.
type CreateMcpConnectionRequest struct {
	McpConnectionRequest
	McpHeaderAuthRequest
}

func (r *CreateMcpConnectionRequest) ValidateAndSanitize() error {
	if err := r.McpConnectionRequest.ValidateAndSanitize(); err != nil {
		return err
	}
	if err := r.McpHeaderAuthRequest.ValidateAndSanitize(); err != nil {
		return err
	}
	return r.McpHeaderAuthRequest.ValidateValues()
}
