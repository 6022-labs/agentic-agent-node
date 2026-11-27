package models

const JsonRpcProtocolBinding = "JSONRPC"

// ExternalAgentCard is the remote agent's A2A v1.0 discovery document.
type ExternalAgentCard struct {
	Name                string                   `json:"name"`
	SupportedInterfaces []ExternalAgentInterface `json:"supportedInterfaces"`
}

type ExternalAgentInterface struct {
	Url             string `json:"url"`
	ProtocolBinding string `json:"protocolBinding"`
	ProtocolVersion string `json:"protocolVersion"`
}

// JsonRpcUrl returns the endpoint advertised for the JSONRPC binding, if any.
func (c *ExternalAgentCard) JsonRpcUrl() string {
	for _, supported := range c.SupportedInterfaces {
		if supported.ProtocolBinding == JsonRpcProtocolBinding {
			return supported.Url
		}
	}
	return ""
}
