package tools

import (
	"context"
	"fmt"

	"github.com/6022-labs/agentic-agent-node/src/agent_reasoning/models"
	tool_interfaces "github.com/6022-labs/agentic-agent-node/src/agent_reasoning_self_tools/tools/interfaces"
	"github.com/6022-labs/agentic-agent-node/src/agent_reasoning_self_tools/utils"
)

// RunCommandTool exposes the command executor as a built-in self tool.
type RunCommandTool struct {
	runCommandExecutor tool_interfaces.RunCommandExecutorInterface
}

func NewRunCommandTool(runCommandExecutor tool_interfaces.RunCommandExecutorInterface) *RunCommandTool {
	return &RunCommandTool{runCommandExecutor: runCommandExecutor}
}

func (t *RunCommandTool) Definition() *models.Tool {
	defaults := t.runCommandExecutor.Defaults()

	return &models.Tool{
		Name:        "run_command",
		Description: "Run a shell command on your own machine and return its output. Every call starts a fresh shell, so no directory or variable survives between calls.",
		InputScheme: utils.ObjectScheme(map[string]models.ToolInputSchemeProperty{
			"command":         utils.StringProperty("The shell command to run."),
			"workdir":         utils.StringProperty(fmt.Sprintf("Directory to run it in. Defaults to %s.", defaults.Workdir)),
			"timeout_seconds": utils.IntegerProperty(fmt.Sprintf("Seconds before the command is killed. Defaults to %d (capped at %d), or %d when background.", defaults.TimeoutSeconds, defaults.MaxTimeoutSeconds, defaults.MaxBackgroundSeconds)),
			"yield_seconds":   utils.IntegerProperty(fmt.Sprintf("Seconds to wait before a still-running command is detached, returning its output so far plus a session id. It still stops at timeout_seconds. Defaults to %d; 0 waits for the command however long it takes.", defaults.YieldSeconds)),
			"background":      utils.BooleanProperty("Return a session id immediately instead of waiting, then poll with read_process_output."),
		}, []string{"command"}),
	}
}

func (t *RunCommandTool) Execute(ctx context.Context, args map[string]any) (map[string]any, error) {
	command, err := utils.StringArg(args, "command")
	if err != nil {
		return nil, err
	}

	return utils.ToToolResult(t.runCommandExecutor.Run(
		ctx,
		command,
		utils.OptionalStringArg(args, "workdir"),
		utils.OptionalIntArg(args, "timeout_seconds"),
		utils.OptionalIntArg(args, "yield_seconds"),
		utils.OptionalBoolArg(args, "background"),
	))
}
