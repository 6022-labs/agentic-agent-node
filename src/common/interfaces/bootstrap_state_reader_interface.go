package interfaces

import "github.com/6022-labs/agentic-agent-node/src/common/models"

// BootstrapStateReaderInterface is the read-only view of the agent's bootstrap
// state. Consumed by the gateway (controllers + middleware).
type BootstrapStateReaderInterface interface {
	Get() models.BootstrapState
}
