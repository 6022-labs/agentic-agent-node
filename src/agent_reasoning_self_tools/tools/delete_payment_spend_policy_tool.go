package tools

import (
	"context"

	"github.com/6022-labs/agentic-agent-node/src/agent_reasoning/models"
	tool_interfaces "github.com/6022-labs/agentic-agent-node/src/agent_reasoning_self_tools/tools/interfaces"
	"github.com/6022-labs/agentic-agent-node/src/agent_reasoning_self_tools/utils"
)

// DeletePaymentSpendPolicyTool exposes the DeletePaymentSpendPolicy use case as a built-in self tool.
type DeletePaymentSpendPolicyTool struct {
	deletePaymentSpendPolicyExecutor tool_interfaces.DeletePaymentSpendPolicyExecutorInterface
}

func NewDeletePaymentSpendPolicyTool(
	deletePaymentSpendPolicyExecutor tool_interfaces.DeletePaymentSpendPolicyExecutorInterface,
) *DeletePaymentSpendPolicyTool {
	return &DeletePaymentSpendPolicyTool{deletePaymentSpendPolicyExecutor: deletePaymentSpendPolicyExecutor}
}

func (t *DeletePaymentSpendPolicyTool) Definition() *models.Tool {
	return &models.Tool{
		Name: "delete_payment_spend_policy",
		Description: "Reset your spending policy. Every rule is removed, leaving you unable to pay " +
			"anything until a rule allows it again.",
		InputScheme: utils.ObjectScheme(map[string]models.ToolInputSchemeProperty{}, nil),
	}
}

func (t *DeletePaymentSpendPolicyTool) Execute(ctx context.Context, _ map[string]any) (map[string]any, error) {
	if err := t.deletePaymentSpendPolicyExecutor.Execute(ctx); err != nil {
		return nil, err
	}
	return utils.SuccessResult(), nil
}

func (t *DeletePaymentSpendPolicyTool) IsEnabledByDefault() bool {
	return false
}

func (t *DeletePaymentSpendPolicyTool) GroupName() string {
	return GroupPayments
}
