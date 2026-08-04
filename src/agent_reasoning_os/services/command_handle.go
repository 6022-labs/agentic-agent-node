package services

import (
	"context"
	"errors"
	"os/exec"
	"sync"
	"syscall"

	"github.com/6022-labs/agentic-agent-node/src/agent_reasoning_os/services/interfaces"
)

// commandHandle owns one running process: its pipes, its buffered output and its exit state.
type commandHandle struct {
	cmd    *exec.Cmd
	stdout *cappedBuffer
	stderr *cappedBuffer
	done   chan struct{}

	mutex    sync.Mutex
	running  bool
	exitCode int
	timedOut bool
}

func (h *commandHandle) Snapshot() *interfaces.CommandResult {
	h.mutex.Lock()
	running, exitCode, timedOut := h.running, h.exitCode, h.timedOut
	h.mutex.Unlock()

	result := &interfaces.CommandResult{
		Stdout:    h.stdout.String(),
		Stderr:    h.stderr.String(),
		Running:   running,
		TimedOut:  timedOut,
		Truncated: h.stdout.Truncated() || h.stderr.Truncated(),
	}
	if !running {
		result.ExitCode = &exitCode
	}
	return result
}

func (h *commandHandle) Done() <-chan struct{} {
	return h.done
}

// The lock is held across the signal: releasing it first lets wait() reap the pid, which the kernel may then reuse.
func (h *commandHandle) Kill() error {
	h.mutex.Lock()
	defer h.mutex.Unlock()

	if !h.running || h.cmd.Process == nil {
		return nil
	}

	// Negative pid targets the whole process group, so pipelines and grandchildren die with the shell.
	if err := syscall.Kill(-h.cmd.Process.Pid, syscall.SIGKILL); err != nil && !errors.Is(err, syscall.ESRCH) {
		return err
	}
	return nil
}

func (h *commandHandle) wait() {
	err := h.cmd.Wait()

	h.mutex.Lock()
	h.running = false
	h.exitCode = exitCodeOf(err)
	h.mutex.Unlock()

	close(h.done)
}

// watch kills the process group when the command's own deadline expires or its owner goes away.
func (h *commandHandle) watch(ctx context.Context) {
	select {
	case <-h.done:
	case <-ctx.Done():
		// Both arms can be ready at once; the command winning that tie did not time out.
		select {
		case <-h.done:
			return
		default:
		}
		if errors.Is(ctx.Err(), context.DeadlineExceeded) {
			h.mutex.Lock()
			h.timedOut = true
			h.mutex.Unlock()
		}
		_ = h.Kill()
	}
}

func exitCodeOf(err error) int {
	if err == nil {
		return 0
	}
	var exitError *exec.ExitError
	if errors.As(err, &exitError) {
		return exitError.ExitCode()
	}
	return -1
}
