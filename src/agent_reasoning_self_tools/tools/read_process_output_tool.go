package tools

import (
	"context"

	"github.com/6022-labs/agentic-agent-node/src/agent_reasoning/models"
	tool_interfaces "github.com/6022-labs/agentic-agent-node/src/agent_reasoning_self_tools/tools/interfaces"
	"github.com/6022-labs/agentic-agent-node/src/agent_reasoning_self_tools/utils"
)

// ReadProcessOutputTool exposes a background command's output as a built-in self tool.
type ReadProcessOutputTool struct {
	readProcessOutputExecutor tool_interfaces.ReadProcessOutputExecutorInterface
}

func NewReadProcessOutputTool(readProcessOutputExecutor tool_interfaces.ReadProcessOutputExecutorInterface) *ReadProcessOutputTool {
	return &ReadProcessOutputTool{readProcessOutputExecutor: readProcessOutputExecutor}
}

func (t *ReadProcessOutputTool) Definition() *models.Tool {
	return &models.Tool{
		Name:        "read_process_output",
		Description: "Read everything a background command has printed so far, plus its exit code once it has finished.",
		InputScheme: utils.ObjectScheme(map[string]models.ToolInputSchemeProperty{
			"session_id": utils.StringProperty("Session id returned by run_command (see list_processes)."),
		}, []string{"session_id"}),
	}
}

func (t *ReadProcessOutputTool) Execute(_ context.Context, args map[string]any) (map[string]any, error) {
	sessionId, err := utils.UuidArg(args, "session_id")
	if err != nil {
		return nil, err
	}

	return utils.ToToolResult(t.readProcessOutputExecutor.ReadOutput(sessionId))
}
