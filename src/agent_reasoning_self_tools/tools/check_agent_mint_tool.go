package tools

import (
	"context"

	"github.com/6022-labs/agentic-agent-node/src/agent_reasoning/models"
	tool_interfaces "github.com/6022-labs/agentic-agent-node/src/agent_reasoning_self_tools/tools/interfaces"
	"github.com/6022-labs/agentic-agent-node/src/agent_reasoning_self_tools/utils"
)

// CheckAgentMintTool exposes the CheckAgentMint use case as a built-in self tool.
type CheckAgentMintTool struct {
	checkAgentMintExecutor tool_interfaces.CheckAgentMintExecutorInterface
}

func NewCheckAgentMintTool(checkAgentMintExecutor tool_interfaces.CheckAgentMintExecutorInterface) *CheckAgentMintTool {
	return &CheckAgentMintTool{checkAgentMintExecutor: checkAgentMintExecutor}
}

func (t *CheckAgentMintTool) Definition() *models.Tool {
	return &models.Tool{
		Name:        "check_mint",
		Description: "Reconcile your mint status from the chain (idempotent read, no transaction).",
		InputScheme: utils.EmptyObjectScheme(),
	}
}

func (t *CheckAgentMintTool) Execute(ctx context.Context, _ map[string]any) (map[string]any, error) {
	return utils.ToToolResult(t.checkAgentMintExecutor.Execute(ctx))
}

func (t *CheckAgentMintTool) IsEnabledByDefault() bool {
	return false
}
