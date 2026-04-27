package interfaces

import (
	common_interfaces "github.com/6022-labs/agentic-agent-node/src/common/interfaces"
	"github.com/6022-labs/agentic-agent-node/src/common/models/enums"
)

// BootstrapStateHolderInterface is the write surface for the wallet bootstrap
// lifecycle. The gateway gets a narrower read-only view via
// common.BootstrapStateReaderInterface, which this interface embeds.
type BootstrapStateHolderInterface interface {
	common_interfaces.BootstrapStateReaderInterface

	SetStatus(status enums.BootstrapStatus)
	SetError(message string)
	SetReady(address string)
}
