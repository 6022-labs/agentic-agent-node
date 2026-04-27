package responses

import "github.com/6022-labs/agentic-agent-node/src/common/models/enums"

type GetBootstrapStatusResponse struct {
	Status       enums.BootstrapStatus `json:"status"`
	AgentAddress string                `json:"agentAddress,omitempty"`
	LastError    string                `json:"lastError,omitempty"`
}
