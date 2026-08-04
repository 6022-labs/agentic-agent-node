package interfaces

import (
	os_interfaces "github.com/6022-labs/agentic-agent-node/src/agent_reasoning_os/services/interfaces"
)

// ListProcessesExecutorInterface lists the detached commands still tracked, oldest first.
type ListProcessesExecutorInterface interface {
	List() []*os_interfaces.BackgroundProcess
}
