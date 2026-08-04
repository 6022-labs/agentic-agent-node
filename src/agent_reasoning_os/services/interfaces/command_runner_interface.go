package interfaces

import (
	"context"
)

// CommandRunnerInterface spawns processes on the machine the runtime lives on.
type CommandRunnerInterface interface {
	// Run blocks until the command exits, or until its timeout kills the whole process group.
	Run(ctx context.Context, spec *CommandSpec) (*CommandResult, error)

	// Start detaches the command from the caller's context — the handle owns its lifetime and its own ceiling.
	Start(spec *CommandSpec) (CommandHandleInterface, error)
}
