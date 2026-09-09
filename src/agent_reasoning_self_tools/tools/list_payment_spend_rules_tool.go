package tools

import (
	"context"

	"github.com/6022-labs/agentic-agent-node/src/agent_reasoning/models"
	tool_interfaces "github.com/6022-labs/agentic-agent-node/src/agent_reasoning_self_tools/tools/interfaces"
	"github.com/6022-labs/agentic-agent-node/src/agent_reasoning_self_tools/utils"
)

// ListPaymentSpendRulesTool exposes the ListPaymentSpendRules use case as a built-in self tool.
type ListPaymentSpendRulesTool struct {
	listPaymentSpendRulesExecutor tool_interfaces.ListPaymentSpendRulesExecutorInterface
}

func NewListPaymentSpendRulesTool(
	listPaymentSpendRulesExecutor tool_interfaces.ListPaymentSpendRulesExecutorInterface,
) *ListPaymentSpendRulesTool {
	return &ListPaymentSpendRulesTool{listPaymentSpendRulesExecutor: listPaymentSpendRulesExecutor}
}

func (t *ListPaymentSpendRulesTool) Definition() *models.Tool {
	return &models.Tool{
		Name: "list_payment_spend_rules",
		Description: "List your own spend rules. A quote you are asked to sign must match a rule on every facet, " +
			"and is capped by the smallest matching maxAmountPerPayment.",
		InputScheme: utils.ObjectScheme(map[string]models.ToolInputSchemeProperty{}, nil),
	}
}

func (t *ListPaymentSpendRulesTool) Execute(ctx context.Context, _ map[string]any) (map[string]any, error) {
	return utils.ToToolResult(t.listPaymentSpendRulesExecutor.Execute(ctx))
}

func (t *ListPaymentSpendRulesTool) IsEnabledByDefault() bool {
	return false
}

func (t *ListPaymentSpendRulesTool) GroupName() string {
	return GroupPayments
}
