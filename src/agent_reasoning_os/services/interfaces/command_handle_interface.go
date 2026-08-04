package interfaces

import ()

// CommandHandleInterface is one running command; the runner owns its pipes and output buffer.
type CommandHandleInterface interface {
	// Snapshot returns the output captured so far, and the exit code once Running is false.
	Snapshot() *CommandResult

	// Done is closed when the command exits, so a caller can wait on it instead of polling Snapshot.
	Done() <-chan struct{}

	// Kill signals the command's whole process group; it is a no-op once the command has exited.
	Kill() error
}
