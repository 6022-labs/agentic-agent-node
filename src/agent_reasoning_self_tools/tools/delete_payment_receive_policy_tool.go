package tools

import (
	"context"

	"github.com/6022-labs/agentic-agent-node/src/agent_reasoning/models"
	tool_interfaces "github.com/6022-labs/agentic-agent-node/src/agent_reasoning_self_tools/tools/interfaces"
	"github.com/6022-labs/agentic-agent-node/src/agent_reasoning_self_tools/utils"
)

// DeletePaymentReceivePolicyTool exposes the DeletePaymentReceivePolicy use case as a built-in self tool.
type DeletePaymentReceivePolicyTool struct {
	deletePaymentReceivePolicyExecutor tool_interfaces.DeletePaymentReceivePolicyExecutorInterface
}

func NewDeletePaymentReceivePolicyTool(
	deletePaymentReceivePolicyExecutor tool_interfaces.DeletePaymentReceivePolicyExecutorInterface,
) *DeletePaymentReceivePolicyTool {
	return &DeletePaymentReceivePolicyTool{deletePaymentReceivePolicyExecutor: deletePaymentReceivePolicyExecutor}
}

func (t *DeletePaymentReceivePolicyTool) Definition() *models.Tool {
	return &models.Tool{
		Name:        "delete_payment_receive_policy",
		Description: "Delete your payment policy; a default free rule is re-materialized on next access.",
		InputScheme: utils.EmptyObjectScheme(),
	}
}

func (t *DeletePaymentReceivePolicyTool) Execute(ctx context.Context, _ map[string]any) (map[string]any, error) {
	if err := t.deletePaymentReceivePolicyExecutor.Execute(ctx); err != nil {
		return nil, err
	}
	return utils.SuccessResult(), nil
}

func (t *DeletePaymentReceivePolicyTool) IsEnabledByDefault() bool {
	return false
}

func (t *DeletePaymentReceivePolicyTool) GroupName() string {
	return GroupPayments
}
