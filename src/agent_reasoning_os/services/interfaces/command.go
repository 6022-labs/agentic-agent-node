package interfaces

import (
	"time"

	"github.com/google/uuid"
)

// CommandDefaults are the configured values a caller gets when it leaves an argument out.
type CommandDefaults struct {
	Workdir              string
	TimeoutSeconds       int
	MaxTimeoutSeconds    int
	MaxBackgroundSeconds int
	YieldSeconds         int
}

// CommandSpec is one command to run, already defaulted and clamped.
type CommandSpec struct {
	Command string
	Workdir string
	Timeout time.Duration
}

// CommandResult is what a command produced so far; a nil ExitCode means it has not exited.
type CommandResult struct {
	SessionId *uuid.UUID `json:"session_id,omitempty"`
	Stdout    string     `json:"stdout"`
	Stderr    string     `json:"stderr"`
	ExitCode  *int       `json:"exit_code,omitempty"`
	Running   bool       `json:"running"`
	TimedOut  bool       `json:"timed_out"`
	Truncated bool       `json:"truncated"`
}

// BackgroundProcess is one detached command tracked by the process registry.
type BackgroundProcess struct {
	Id        uuid.UUID  `json:"session_id"`
	Command   string     `json:"command"`
	Workdir   string     `json:"workdir"`
	StartedAt time.Time  `json:"started_at"`
	EndedAt   *time.Time `json:"ended_at,omitempty"`
	Running   bool       `json:"running"`
	ExitCode  *int       `json:"exit_code,omitempty"`
}
