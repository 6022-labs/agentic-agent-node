package interfaces

import (
	"context"

	os_interfaces "github.com/6022-labs/agentic-agent-node/src/agent_reasoning_os/services/interfaces"
)

// RunCommandExecutorInterface runs a command on the agent's own machine, or detaches it and reports the session id.
type RunCommandExecutorInterface interface {
	Run(
		ctx context.Context,
		command string,
		workdir string,
		timeoutSeconds *int,
		yieldSeconds *int,
		background *bool,
	) (*os_interfaces.CommandResult, error)
	// Defaults reports what an omitted argument resolves to, so the tool description matches the running config.
	Defaults() *os_interfaces.CommandDefaults
}
