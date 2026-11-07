package services

import (
	"sync"
	"time"

	"github.com/6022-labs/agentic-agent-node/src/agent_reasoning_os/services/interfaces"
	"github.com/6022-labs/agentic-agent-node/src/agent_reasoning_os/settings"
	"github.com/6022-labs/agentic-agent-node/src/common/errors"
	"github.com/google/uuid"
)

// ProcessRegistry keeps detached commands in memory for the runtime's lifetime; nothing survives a restart.
type ProcessRegistry struct {
	mutex           sync.Mutex
	commandSettings *settings.CommandSettings
	entries         map[uuid.UUID]*processEntry
}

type processEntry struct {
	process *interfaces.BackgroundProcess
	handle  interfaces.CommandHandleInterface
}

func NewProcessRegistry(commandSettings *settings.CommandSettings) *ProcessRegistry {
	return &ProcessRegistry{
		commandSettings: commandSettings,
		entries:         map[uuid.UUID]*processEntry{},
	}
}

func (r *ProcessRegistry) Register(
	command string,
	workdir string,
	handle interfaces.CommandHandleInterface,
) (*interfaces.BackgroundProcess, error) {
	r.mutex.Lock()
	defer r.mutex.Unlock()

	r.purge()

	running := 0
	for _, entry := range r.entries {
		if entry.process.Running {
			running++
		}
	}
	if running >= r.commandSettings.MaxBackgroundProcesses {
		return nil, errors.NewConflictError(
			"too_many_background_processes",
			"the maximum number of concurrent background processes is already running",
		)
	}

	process := &interfaces.BackgroundProcess{
		Id:        uuid.New(),
		Command:   command,
		Workdir:   workdir,
		StartedAt: time.Now().UTC(),
		Running:   true,
	}
	r.entries[process.Id] = &processEntry{process: process, handle: handle}

	return r.snapshot(r.entries[process.Id]), nil
}

func (r *ProcessRegistry) List() []*interfaces.BackgroundProcess {
	r.mutex.Lock()
	defer r.mutex.Unlock()

	r.purge()

	processes := make([]*interfaces.BackgroundProcess, 0, len(r.entries))
	for _, entry := range r.entries {
		processes = append(processes, r.snapshot(entry))
	}
	return processes
}

func (r *ProcessRegistry) Get(id uuid.UUID) (*interfaces.BackgroundProcess, *interfaces.CommandResult, error) {
	r.mutex.Lock()
	defer r.mutex.Unlock()

	entry, found := r.entries[id]
	if !found {
		return nil, nil, r.notFound(id)
	}
	return r.snapshot(entry), entry.handle.Snapshot(), nil
}

func (r *ProcessRegistry) Kill(id uuid.UUID) error {
	r.mutex.Lock()
	defer r.mutex.Unlock()

	entry, found := r.entries[id]
	if !found {
		return r.notFound(id)
	}
	if err := entry.handle.Kill(); err != nil {
		return errors.NewInternalError("kill_failed", err)
	}
	r.refresh(entry)
	return nil
}

func (*ProcessRegistry) notFound(id uuid.UUID) error {
	return errors.NewNotFoundError("process_not_found", "no background process with id "+id.String())
}

// snapshot hands out a copy: the stored row keeps being mutated by refresh, under a lock the caller does not hold.
func (r *ProcessRegistry) snapshot(entry *processEntry) *interfaces.BackgroundProcess {
	process := *r.refresh(entry)
	if process.EndedAt != nil {
		endedAt := *process.EndedAt
		process.EndedAt = &endedAt
	}
	if process.ExitCode != nil {
		exitCode := *process.ExitCode
		process.ExitCode = &exitCode
	}
	return &process
}

// refresh folds the handle's live state back into the process row; the handle is the source of truth.
func (*ProcessRegistry) refresh(entry *processEntry) *interfaces.BackgroundProcess {
	result := entry.handle.Snapshot()
	if entry.process.Running && !result.Running {
		endedAt := time.Now().UTC()
		entry.process.Running = false
		entry.process.EndedAt = &endedAt
		if result.ExitCode != nil {
			exitCode := *result.ExitCode
			entry.process.ExitCode = &exitCode
		}
	}
	return entry.process
}

// purge runs opportunistically, not on a timer: the per-command ceiling is enforced by the handle itself.
func (r *ProcessRegistry) purge() {
	cutoff := time.Now().UTC().Add(-time.Duration(r.commandSettings.FinishedRetentionSeconds) * time.Second)
	for id, entry := range r.entries {
		r.refresh(entry)
		if entry.process.EndedAt != nil && entry.process.EndedAt.Before(cutoff) {
			delete(r.entries, id)
		}
	}
}
