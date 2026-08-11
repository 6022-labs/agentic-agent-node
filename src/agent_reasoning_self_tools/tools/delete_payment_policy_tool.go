package tools

import (
	"context"

	"github.com/6022-labs/agentic-agent-node/src/agent_reasoning/models"
	tool_interfaces "github.com/6022-labs/agentic-agent-node/src/agent_reasoning_self_tools/tools/interfaces"
	"github.com/6022-labs/agentic-agent-node/src/agent_reasoning_self_tools/utils"
)

// DeletePaymentPolicyTool exposes the DeletePaymentPolicy use case as a built-in self tool.
type DeletePaymentPolicyTool struct {
	deletePaymentPolicyExecutor tool_interfaces.DeletePaymentPolicyExecutorInterface
}

func NewDeletePaymentPolicyTool(deletePaymentPolicyExecutor tool_interfaces.DeletePaymentPolicyExecutorInterface) *DeletePaymentPolicyTool {
	return &DeletePaymentPolicyTool{deletePaymentPolicyExecutor: deletePaymentPolicyExecutor}
}

func (t *DeletePaymentPolicyTool) Definition() *models.Tool {
	return &models.Tool{
		Name:        "delete_payment_policy",
		Description: "Delete your payment policy; a default free rule is re-materialized on next access.",
		InputScheme: utils.EmptyObjectScheme(),
	}
}

func (t *DeletePaymentPolicyTool) Execute(ctx context.Context, _ map[string]any) (map[string]any, error) {
	if err := t.deletePaymentPolicyExecutor.Execute(ctx); err != nil {
		return nil, err
	}
	return utils.SuccessResult(), nil
}

func (t *DeletePaymentPolicyTool) IsEnabledByDefault() bool {
	return false
}
