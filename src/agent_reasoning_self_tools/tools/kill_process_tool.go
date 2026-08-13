package tools

import (
	"context"

	"github.com/6022-labs/agentic-agent-node/src/agent_reasoning/models"
	tool_interfaces "github.com/6022-labs/agentic-agent-node/src/agent_reasoning_self_tools/tools/interfaces"
	"github.com/6022-labs/agentic-agent-node/src/agent_reasoning_self_tools/utils"
)

// KillProcessTool exposes background process termination as a built-in self tool.
type KillProcessTool struct {
	killProcessExecutor tool_interfaces.KillProcessExecutorInterface
}

func NewKillProcessTool(killProcessExecutor tool_interfaces.KillProcessExecutorInterface) *KillProcessTool {
	return &KillProcessTool{killProcessExecutor: killProcessExecutor}
}

func (t *KillProcessTool) Definition() *models.Tool {
	return &models.Tool{
		Name: "kill_process",
		Description: "Stop a background command you started here, along with anything it spawned. " +
			"It takes a session id, not a process id from ps.",
		InputScheme: utils.ObjectScheme(map[string]models.ToolInputSchemeProperty{
			"session_id": utils.StringProperty("Session id returned by run_command (see list_processes)."),
		}, []string{"session_id"}),
	}
}

func (t *KillProcessTool) Execute(ctx context.Context, args map[string]any) (map[string]any, error) {
	sessionId, err := utils.UuidArg(args, "session_id")
	if err != nil {
		return nil, err
	}
	if err := t.killProcessExecutor.Kill(ctx, sessionId); err != nil {
		return nil, err
	}
	return utils.SuccessResult(), nil
}

func (t *KillProcessTool) IsEnabledByDefault() bool {
	return true
}

func (t *KillProcessTool) GroupName() string {
	return GroupMachine
}
