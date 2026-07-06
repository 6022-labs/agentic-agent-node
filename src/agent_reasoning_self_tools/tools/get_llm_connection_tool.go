package tools

import (
	"context"
	"github.com/6022-labs/agentic-agent-node/src/agent_reasoning/models"
	tool_interfaces "github.com/6022-labs/agentic-agent-node/src/agent_reasoning_self_tools/tools/interfaces"
	"github.com/6022-labs/agentic-agent-node/src/agent_reasoning_self_tools/utils"
)

// GetLLMConnectionTool exposes the GetLLMConnection use case as a built-in self tool.
type GetLLMConnectionTool struct {
	getLLMConnectionExecutor tool_interfaces.GetLLMConnectionExecutorInterface
}

func NewGetLLMConnectionTool(getLLMConnectionExecutor tool_interfaces.GetLLMConnectionExecutorInterface) *GetLLMConnectionTool {
	return &GetLLMConnectionTool{getLLMConnectionExecutor: getLLMConnectionExecutor}
}

func (t *GetLLMConnectionTool) Definition() *models.Tool {
	return &models.Tool{
		Name:        "get_llm_connection",
		Description: "Read your configured LLM connection (platform, model, base url). The api key is never returned.",
		InputScheme: utils.EmptyObjectScheme(),
	}
}

func (t *GetLLMConnectionTool) Execute(ctx context.Context, _ map[string]any) (map[string]any, error) {
	return utils.ToToolResult(t.getLLMConnectionExecutor.Execute(ctx))
}
