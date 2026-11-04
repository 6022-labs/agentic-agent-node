package tools

import (
	"context"

	"github.com/6022-labs/agentic-agent-node/src/agent_reasoning/models"
	tool_interfaces "github.com/6022-labs/agentic-agent-node/src/agent_reasoning_self_tools/tools/interfaces"
	"github.com/6022-labs/agentic-agent-node/src/agent_reasoning_self_tools/utils"
)

// GetPaymentSpendPolicyTool exposes the GetPaymentSpendPolicy use case as a built-in self tool.
type GetPaymentSpendPolicyTool struct {
	getPaymentSpendPolicyExecutor tool_interfaces.GetPaymentSpendPolicyExecutorInterface
}

func NewGetPaymentSpendPolicyTool(
	getPaymentSpendPolicyExecutor tool_interfaces.GetPaymentSpendPolicyExecutorInterface,
) *GetPaymentSpendPolicyTool {
	return &GetPaymentSpendPolicyTool{getPaymentSpendPolicyExecutor: getPaymentSpendPolicyExecutor}
}

func (t *GetPaymentSpendPolicyTool) Definition() *models.Tool {
	return &models.Tool{
		Name:        "get_payment_spend_policy",
		Description: "Read your own spending policy: the payees you may pay, whatever the token.",
		InputScheme: utils.ObjectScheme(map[string]models.ToolInputSchemeProperty{}, nil),
	}
}

func (t *GetPaymentSpendPolicyTool) Execute(ctx context.Context, _ map[string]any) (map[string]any, error) {
	return utils.ToToolResult(t.getPaymentSpendPolicyExecutor.Execute(ctx))
}

func (t *GetPaymentSpendPolicyTool) IsEnabledByDefault() bool {
	return false
}

func (t *GetPaymentSpendPolicyTool) GroupName() string {
	return GroupPayments
}
