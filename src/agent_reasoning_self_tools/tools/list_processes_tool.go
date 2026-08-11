package tools

import (
	"context"

	"github.com/6022-labs/agentic-agent-node/src/agent_reasoning/models"
	tool_interfaces "github.com/6022-labs/agentic-agent-node/src/agent_reasoning_self_tools/tools/interfaces"
	"github.com/6022-labs/agentic-agent-node/src/agent_reasoning_self_tools/utils"
)

// ListProcessesTool exposes the tracked background processes as a built-in self tool.
type ListProcessesTool struct {
	listProcessesExecutor tool_interfaces.ListProcessesExecutorInterface
}

func NewListProcessesTool(listProcessesExecutor tool_interfaces.ListProcessesExecutorInterface) *ListProcessesTool {
	return &ListProcessesTool{listProcessesExecutor: listProcessesExecutor}
}

func (t *ListProcessesTool) Definition() *models.Tool {
	return &models.Tool{
		Name:        "list_processes",
		Description: "List the background commands you started here: their session id, command, and whether they are still running. It never lists the other processes on this machine, and it forgets everything when the runtime restarts.",
		InputScheme: utils.EmptyObjectScheme(),
	}
}

func (t *ListProcessesTool) Execute(_ context.Context, _ map[string]any) (map[string]any, error) {
	// Wrapped in an object: a tool result is never a bare array.
	return utils.ToToolResult(map[string]any{"processes": t.listProcessesExecutor.List()}, nil)
}

func (t *ListProcessesTool) IsEnabledByDefault() bool {
	return true
}

func (t *ListProcessesTool) GroupName() string {
	return GroupMachine
}
