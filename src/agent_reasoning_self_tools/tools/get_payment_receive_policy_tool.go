package tools

import (
	"context"

	"github.com/6022-labs/agentic-agent-node/src/agent_reasoning/models"
	tool_interfaces "github.com/6022-labs/agentic-agent-node/src/agent_reasoning_self_tools/tools/interfaces"
	"github.com/6022-labs/agentic-agent-node/src/agent_reasoning_self_tools/utils"
)

// GetPaymentReceivePolicyTool exposes the GetPaymentReceivePolicy use case as a built-in self tool.
type GetPaymentReceivePolicyTool struct {
	getPaymentReceivePolicyExecutor tool_interfaces.GetPaymentReceivePolicyExecutorInterface
}

func NewGetPaymentReceivePolicyTool(
	getPaymentReceivePolicyExecutor tool_interfaces.GetPaymentReceivePolicyExecutorInterface,
) *GetPaymentReceivePolicyTool {
	return &GetPaymentReceivePolicyTool{getPaymentReceivePolicyExecutor: getPaymentReceivePolicyExecutor}
}

func (t *GetPaymentReceivePolicyTool) Definition() *models.Tool {
	return &models.Tool{
		Name:        "get_payment_receive_policy",
		Description: "Read your payment policy head.",
		InputScheme: utils.EmptyObjectScheme(),
	}
}

func (t *GetPaymentReceivePolicyTool) Execute(ctx context.Context, _ map[string]any) (map[string]any, error) {
	return utils.ToToolResult(t.getPaymentReceivePolicyExecutor.Execute(ctx))
}

func (t *GetPaymentReceivePolicyTool) IsEnabledByDefault() bool {
	return false
}

func (t *GetPaymentReceivePolicyTool) GroupName() string {
	return GroupPayments
}
