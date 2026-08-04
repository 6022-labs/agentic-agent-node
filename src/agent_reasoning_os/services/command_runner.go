package services

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"os/user"
	"strconv"
	"syscall"
	"time"

	"github.com/6022-labs/agentic-agent-node/src/agent_reasoning_os/services/interfaces"
	"github.com/6022-labs/agentic-agent-node/src/agent_reasoning_os/settings"
	"go.uber.org/zap"
)

const fallbackShellPath = "/bin/sh"

// pipeDrainDelay bounds how long Wait keeps reading after the process itself has gone.
const pipeDrainDelay = 5 * time.Second

// CommandRunner is the only place in the codebase that spawns processes.
type CommandRunner struct {
	logger          *zap.Logger
	osSettings      *settings.OsSettings
	commandSettings *settings.CommandSettings
	shellPath       string
	credential      *syscall.Credential
}

func NewCommandRunner(logger *zap.Logger, osSettings *settings.OsSettings, commandSettings *settings.CommandSettings) (*CommandRunner, error) {
	credential, err := commandCredential(osSettings.CommandUser)
	if err != nil {
		return nil, err
	}

	shellPath, err := exec.LookPath("bash")
	if err != nil {
		logger.Warn("bash not found, falling back to sh", zap.String("shellPath", fallbackShellPath))
		shellPath = fallbackShellPath
	}

	if err := os.MkdirAll(osSettings.Workdir, 0o755); err != nil {
		logger.Warn("failed to create the default workdir",
			zap.String("workdir", osSettings.Workdir),
			zap.Error(err),
		)
	} else if credential != nil {
		if err := os.Chown(osSettings.Workdir, int(credential.Uid), int(credential.Gid)); err != nil {
			logger.Warn("failed to give the command user its workdir, so commands may not be able to write",
				zap.String("workdir", osSettings.Workdir),
				zap.String("commandUser", osSettings.CommandUser),
				zap.Error(err),
			)
		}
	}

	if credential == nil {
		logger.Warn("commands run as the runtime's own user, so the agent can signal the runtime, the gateway and nginx")
	}

	return &CommandRunner{
		logger:          logger,
		osSettings:      osSettings,
		commandSettings: commandSettings,
		shellPath:       shellPath,
		credential:      credential,
	}, nil
}

func (r *CommandRunner) Run(ctx context.Context, spec *interfaces.CommandSpec) (*interfaces.CommandResult, error) {
	runCtx, cancel := context.WithTimeout(ctx, spec.Timeout)
	defer cancel()

	handle, err := r.start(runCtx, spec)
	if err != nil {
		return nil, err
	}

	<-handle.done

	return handle.Snapshot(), nil
}

func (r *CommandRunner) Start(spec *interfaces.CommandSpec) (interfaces.CommandHandleInterface, error) {
	// Rooted in Background, not the caller's context: the http request that started it is gone long before the command is.
	ctx, cancel := context.WithTimeout(context.Background(), spec.Timeout)

	handle, err := r.start(ctx, spec)
	if err != nil {
		cancel()
		return nil, err
	}

	go func() {
		<-handle.done
		cancel()
	}()

	return handle, nil
}

func (r *CommandRunner) start(ctx context.Context, spec *interfaces.CommandSpec) (*commandHandle, error) {
	cmd := exec.Command(r.shellPath, "-c", spec.Command)
	cmd.Dir = spec.Workdir
	cmd.Env = r.commandEnv()
	cmd.SysProcAttr = &syscall.SysProcAttr{Setpgid: true, Credential: r.credential}
	// A grandchild that escapes the process group keeps the output pipe open, and Wait would otherwise block on it forever.
	cmd.WaitDelay = pipeDrainDelay

	handle := &commandHandle{
		cmd:     cmd,
		stdout:  newCappedBuffer(r.commandSettings.MaxOutputChars),
		stderr:  newCappedBuffer(r.commandSettings.MaxOutputChars),
		done:    make(chan struct{}),
		running: true,
	}
	cmd.Stdout = handle.stdout
	cmd.Stderr = handle.stderr

	if err := cmd.Start(); err != nil {
		return nil, err
	}

	go handle.wait()
	go handle.watch(ctx)

	return handle, nil
}

// commandCredential resolves the account commands drop to; an empty name keeps the runtime's own, which can kill it.
func commandCredential(name string) (*syscall.Credential, error) {
	if name == "" {
		return nil, nil
	}

	account, err := user.Lookup(name)
	if err != nil {
		return nil, fmt.Errorf("os.command_user %q cannot be resolved: %w", name, err)
	}

	uid, err := strconv.Atoi(account.Uid)
	if err != nil {
		return nil, fmt.Errorf("os.command_user %q has a non-numeric uid %q", name, account.Uid)
	}
	gid, err := strconv.Atoi(account.Gid)
	if err != nil {
		return nil, fmt.Errorf("os.command_user %q has a non-numeric gid %q", name, account.Gid)
	}

	// NoSetGroups: setgroups is refused unless the runtime is root, and the account's own group is all a command needs.
	return &syscall.Credential{Uid: uint32(uid), Gid: uint32(gid), NoSetGroups: true}, nil
}

// Not os.Environ(): hygiene only, and /proc/1/environ still exposes the vault token whenever commands run as root.
func (r *CommandRunner) commandEnv() []string {
	return []string{
		"PATH=/usr/local/sbin:/usr/local/bin:/usr/sbin:/usr/bin:/sbin:/bin",
		"HOME=" + r.osSettings.Workdir,
		"TERM=dumb",
		"LANG=C.UTF-8",
	}
}
