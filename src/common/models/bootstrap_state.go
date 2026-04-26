package models

import "github.com/6022-labs/agentic-agent-node/src/common/models/enums"

// BootstrapState describes the wallet-loading lifecycle. Identity (NFT,
// collection, ens) is resolved separately and lazily by AgentIdentityResolver.
type BootstrapState struct {
	Status       enums.BootstrapStatus
	AgentAddress string
	LastError    string
}
