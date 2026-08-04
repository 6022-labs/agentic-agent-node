package interfaces

import (
	"github.com/google/uuid"
)

// ProcessRegistryInterface tracks detached commands for the lifetime of the runtime process.
type ProcessRegistryInterface interface {
	// Register takes ownership of a started command; it fails with a ConflictError past the concurrency ceiling.
	Register(command string, workdir string, handle CommandHandleInterface) (*BackgroundProcess, error)

	// List returns running processes plus finished ones still inside the retention window.
	List() []*BackgroundProcess

	// Get returns the process refreshed from its handle, plus the latest command output.
	Get(id uuid.UUID) (*BackgroundProcess, *CommandResult, error)

	// Kill terminates a process group; killing an already-finished process is not an error.
	Kill(id uuid.UUID) error
}
