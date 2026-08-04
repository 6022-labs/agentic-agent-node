package interfaces

import (
	os_interfaces "github.com/6022-labs/agentic-agent-node/src/agent_reasoning_os/services/interfaces"
	"github.com/google/uuid"
)

// ReadProcessOutputExecutorInterface returns a detached command's output so far.
type ReadProcessOutputExecutorInterface interface {
	ReadOutput(id uuid.UUID) (*os_interfaces.CommandResult, error)
}
