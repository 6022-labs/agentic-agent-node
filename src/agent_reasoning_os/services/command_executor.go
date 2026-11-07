package services

import (
	"context"
	goerrors "errors"
	"io/fs"
	"sort"
	"strings"
	"time"

	metrics_interfaces "github.com/6022-labs/agentic-agent-node/src/agent_reasoning_os/metrics/interfaces"
	"github.com/6022-labs/agentic-agent-node/src/agent_reasoning_os/services/interfaces"
	"github.com/6022-labs/agentic-agent-node/src/agent_reasoning_os/settings"
	"github.com/6022-labs/agentic-agent-node/src/common/errors"
	"github.com/google/uuid"
	"github.com/samber/lo"
)

// CommandExecutor turns loose tool arguments into a bounded run and tracks the detached ones.
type CommandExecutor struct {
	commandRunner   interfaces.CommandRunnerInterface
	processRegistry interfaces.ProcessRegistryInterface
	osSettings      *settings.OsSettings
	commandSettings *settings.CommandSettings
	commandMetrics  metrics_interfaces.CommandMetricsInterface
}

func NewCommandExecutor(
	commandRunner interfaces.CommandRunnerInterface,
	processRegistry interfaces.ProcessRegistryInterface,
	osSettings *settings.OsSettings,
	commandSettings *settings.CommandSettings,
	commandMetrics metrics_interfaces.CommandMetricsInterface,
) *CommandExecutor {
	return &CommandExecutor{
		commandRunner:   commandRunner,
		processRegistry: processRegistry,
		osSettings:      osSettings,
		commandSettings: commandSettings,
		commandMetrics:  commandMetrics,
	}
}

func (s *CommandExecutor) Defaults() *interfaces.CommandDefaults {
	return &interfaces.CommandDefaults{
		Workdir:              s.osSettings.Workdir,
		TimeoutSeconds:       s.commandSettings.DefaultTimeoutSeconds,
		MaxTimeoutSeconds:    s.commandSettings.MaxTimeoutSeconds,
		MaxBackgroundSeconds: s.commandSettings.MaxBackgroundSeconds,
		YieldSeconds:         s.commandSettings.DefaultYieldSeconds,
	}
}

func (s *CommandExecutor) Run(
	ctx context.Context,
	command string,
	workdir string,
	timeoutSeconds *int,
	yieldSeconds *int,
	background *bool,
) (*interfaces.CommandResult, error) {
	if strings.TrimSpace(command) == "" {
		return nil, errors.NewValidationError("command", "must not be empty")
	}
	if timeoutSeconds != nil && *timeoutSeconds <= 0 {
		return nil, errors.NewValidationError("timeout_seconds", "must be positive")
	}
	if yieldSeconds != nil && *yieldSeconds < 0 {
		return nil, errors.NewValidationError("yield_seconds", "must not be negative")
	}

	detached := lo.FromPtr(background)

	spec := &interfaces.CommandSpec{
		Command: command,
		Workdir: lo.Ternary(workdir == "", s.osSettings.Workdir, workdir),
		Timeout: s.resolveTimeout(timeoutSeconds, detached),
	}

	if detached {
		return s.startBackground(ctx, spec)
	}

	if yield := s.resolveYield(yieldSeconds, spec.Timeout); yield > 0 {
		return s.runUntilYield(ctx, spec, yield)
	}

	result, err := s.commandRunner.Run(ctx, spec)
	if err != nil {
		return nil, s.runnerError(err)
	}

	return result, nil
}

func (s *CommandExecutor) List() []*interfaces.BackgroundProcess {
	processes := s.processRegistry.List()
	sort.Slice(processes, func(i, j int) bool { return processes[i].StartedAt.Before(processes[j].StartedAt) })
	return processes
}

func (s *CommandExecutor) ReadOutput(id uuid.UUID) (*interfaces.CommandResult, error) {
	if id == uuid.Nil {
		return nil, errors.NewValidationError("session_id", "must not be empty")
	}

	process, result, err := s.processRegistry.Get(id)
	if err != nil {
		return nil, err
	}

	result.SessionId = &process.Id
	return result, nil
}

func (s *CommandExecutor) Kill(ctx context.Context, id uuid.UUID) error {
	if id == uuid.Nil {
		return errors.NewValidationError("session_id", "must not be empty")
	}

	if err := s.processRegistry.Kill(id); err != nil {
		return err
	}

	s.commandMetrics.RecordProcessKilled(ctx)

	return nil
}

func (s *CommandExecutor) startBackground(
	ctx context.Context,
	spec *interfaces.CommandSpec,
) (*interfaces.CommandResult, error) {
	handle, err := s.commandRunner.Start(spec)
	if err != nil {
		return nil, s.runnerError(err)
	}

	result, err := s.detach(ctx, spec, handle)
	if err != nil {
		// The command is already running; the registry refusing it must not leave an untracked process behind.
		_ = handle.Kill()
		return nil, err
	}

	return result, nil
}

// runUntilYield returns whatever a slow command printed so far and hands back a session id instead of blocking.
func (s *CommandExecutor) runUntilYield(
	ctx context.Context,
	spec *interfaces.CommandSpec,
	yield time.Duration,
) (*interfaces.CommandResult, error) {
	handle, err := s.commandRunner.Start(spec)
	if err != nil {
		return nil, s.runnerError(err)
	}

	timer := time.NewTimer(yield)
	defer timer.Stop()

	select {
	case <-handle.Done():
		return handle.Snapshot(), nil
	case <-timer.C:
	case <-ctx.Done():
	}

	result, err := s.detach(ctx, spec, handle)
	if err != nil {
		// Untracked and unreachable otherwise, so wait it out here rather than orphan it at its own deadline.
		<-handle.Done()
		return handle.Snapshot(), nil
	}

	return result, nil
}

// detach hands a running command to the registry; the snapshot is taken after, since it may already have exited.
func (s *CommandExecutor) detach(
	ctx context.Context,
	spec *interfaces.CommandSpec,
	handle interfaces.CommandHandleInterface,
) (*interfaces.CommandResult, error) {
	process, err := s.processRegistry.Register(spec.Command, spec.Workdir, handle)
	if err != nil {
		return nil, err
	}

	s.commandMetrics.RecordBackgroundStarted(ctx)

	result := handle.Snapshot()
	result.SessionId = &process.Id

	return result, nil
}

// runnerError reports a missing workdir as caller input; a missing shell is also fs.ErrNotExist, so the op is checked.
func (s *CommandExecutor) runnerError(err error) error {
	var pathError *fs.PathError
	if goerrors.As(err, &pathError) && pathError.Op == "chdir" {
		return errors.NewValidationError("workdir", "does not exist")
	}
	return errors.NewInternalError("command_failed", err)
}

// Seconds are compared before conversion: the multiplication overflows int64 well below what a model may send.
func (s *CommandExecutor) resolveTimeout(timeoutSeconds *int, background bool) time.Duration {
	ceiling := time.Duration(s.commandSettings.MaxTimeoutSeconds) * time.Second
	if background {
		ceiling = time.Duration(s.commandSettings.MaxBackgroundSeconds) * time.Second
	}

	if timeoutSeconds == nil {
		if background {
			return ceiling
		}
		return time.Duration(s.commandSettings.DefaultTimeoutSeconds) * time.Second
	}
	if *timeoutSeconds >= int(ceiling/time.Second) {
		return ceiling
	}
	return time.Duration(*timeoutSeconds) * time.Second
}

// A yield reaching the deadline never fires, so it is dropped: the command is killed there anyway.
func (s *CommandExecutor) resolveYield(yieldSeconds *int, timeout time.Duration) time.Duration {
	seconds := lo.FromPtrOr(yieldSeconds, s.commandSettings.DefaultYieldSeconds)
	if seconds <= 0 || seconds >= int(timeout/time.Second) {
		return 0
	}
	return time.Duration(seconds) * time.Second
}
