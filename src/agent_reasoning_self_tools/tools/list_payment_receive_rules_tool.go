package tools

import (
	"context"

	"github.com/6022-labs/agentic-agent-node/src/agent_reasoning/models"
	tool_interfaces "github.com/6022-labs/agentic-agent-node/src/agent_reasoning_self_tools/tools/interfaces"
	"github.com/6022-labs/agentic-agent-node/src/agent_reasoning_self_tools/utils"
)

// ListPaymentReceiveRulesTool exposes the ListPaymentReceiveRules use case as a built-in self tool.
type ListPaymentReceiveRulesTool struct {
	listPaymentReceiveRulesExecutor tool_interfaces.ListPaymentReceiveRulesExecutorInterface
}

func NewListPaymentReceiveRulesTool(
	listPaymentReceiveRulesExecutor tool_interfaces.ListPaymentReceiveRulesExecutorInterface,
) *ListPaymentReceiveRulesTool {
	return &ListPaymentReceiveRulesTool{listPaymentReceiveRulesExecutor: listPaymentReceiveRulesExecutor}
}

func (t *ListPaymentReceiveRulesTool) Definition() *models.Tool {
	return &models.Tool{
		Name:        "list_payment_receive_rules",
		Description: "List your pricing rules for callers: one per token, each with its prices and who pays them.",
		InputScheme: utils.EmptyObjectScheme(),
	}
}

func (t *ListPaymentReceiveRulesTool) Execute(ctx context.Context, _ map[string]any) (map[string]any, error) {
	return utils.ToToolResult(t.listPaymentReceiveRulesExecutor.Execute(ctx))
}

func (t *ListPaymentReceiveRulesTool) IsEnabledByDefault() bool {
	return false
}

func (t *ListPaymentReceiveRulesTool) GroupName() string {
	return GroupPayments
}
