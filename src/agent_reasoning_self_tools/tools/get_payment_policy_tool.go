package tools

import (
	"context"
	"github.com/6022-labs/agentic-agent-node/src/agent_reasoning/models"
	tool_interfaces "github.com/6022-labs/agentic-agent-node/src/agent_reasoning_self_tools/tools/interfaces"
	"github.com/6022-labs/agentic-agent-node/src/agent_reasoning_self_tools/utils"
)

// GetPaymentPolicyTool exposes the GetPaymentPolicy use case as a built-in self tool.
type GetPaymentPolicyTool struct {
	getPaymentPolicyExecutor tool_interfaces.GetPaymentPolicyExecutorInterface
}

func NewGetPaymentPolicyTool(getPaymentPolicyExecutor tool_interfaces.GetPaymentPolicyExecutorInterface) *GetPaymentPolicyTool {
	return &GetPaymentPolicyTool{getPaymentPolicyExecutor: getPaymentPolicyExecutor}
}

func (t *GetPaymentPolicyTool) Definition() *models.Tool {
	return &models.Tool{
		Name:        "get_payment_policy",
		Description: "Read your payment policy head.",
		InputScheme: utils.EmptyObjectScheme(),
	}
}

func (t *GetPaymentPolicyTool) Execute(ctx context.Context, _ map[string]any) (map[string]any, error) {
	return utils.ToToolResult(t.getPaymentPolicyExecutor.Execute(ctx))
}
