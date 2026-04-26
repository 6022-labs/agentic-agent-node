package responses

import "github.com/6022-labs/agentic-agent-node/src/common/models/enums"

type GetBootstrapStatusResponse struct {
	Status       enums.BootstrapStatus    `json:"status"`
	AgentAddress string                   `json:"agentAddress,omitempty"`
	LastError    string                   `json:"lastError,omitempty"`
	Identity     *AgentIdentitySnapshot   `json:"identity,omitempty"`
}

// AgentIdentitySnapshot reports a best-effort identity attempt at the time
// /bootstrap/status was hit. Resolved=true means an NFT exists for this
// agent's wallet; false with Error populated means none yet.
type AgentIdentitySnapshot struct {
	Resolved          bool   `json:"resolved"`
	CollectionAddress string `json:"collectionAddress,omitempty"`
	TokenId           uint64 `json:"tokenId,omitempty"`
	Name              string `json:"name,omitempty"`
	Error             string `json:"error,omitempty"`
}
