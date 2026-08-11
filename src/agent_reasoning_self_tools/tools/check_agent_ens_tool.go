package tools

import (
	"context"

	"github.com/6022-labs/agentic-agent-node/src/agent_reasoning/models"
	tool_interfaces "github.com/6022-labs/agentic-agent-node/src/agent_reasoning_self_tools/tools/interfaces"
	"github.com/6022-labs/agentic-agent-node/src/agent_reasoning_self_tools/utils"
)

// CheckAgentEnsTool exposes the CheckAgentEns use case as a built-in self tool.
type CheckAgentEnsTool struct {
	checkAgentEnsExecutor tool_interfaces.CheckAgentEnsExecutorInterface
}

func NewCheckAgentEnsTool(checkAgentEnsExecutor tool_interfaces.CheckAgentEnsExecutorInterface) *CheckAgentEnsTool {
	return &CheckAgentEnsTool{checkAgentEnsExecutor: checkAgentEnsExecutor}
}

func (t *CheckAgentEnsTool) Definition() *models.Tool {
	return &models.Tool{
		Name:        "check_ens",
		Description: "Reconcile your ENS provisioning status from the chain (idempotent read, no transaction).",
		InputScheme: utils.EmptyObjectScheme(),
	}
}

func (t *CheckAgentEnsTool) Execute(ctx context.Context, _ map[string]any) (map[string]any, error) {
	return utils.ToToolResult(t.checkAgentEnsExecutor.Execute(ctx))
}

func (t *CheckAgentEnsTool) IsEnabledByDefault() bool {
	return false
}

func (t *CheckAgentEnsTool) GroupName() string {
	return GroupIdentity
}
