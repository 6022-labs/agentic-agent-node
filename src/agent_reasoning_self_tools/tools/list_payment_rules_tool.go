package tools

import (
	"context"

	"github.com/6022-labs/agentic-agent-node/src/agent_reasoning/models"
	tool_interfaces "github.com/6022-labs/agentic-agent-node/src/agent_reasoning_self_tools/tools/interfaces"
	"github.com/6022-labs/agentic-agent-node/src/agent_reasoning_self_tools/utils"
)

// ListPaymentRulesTool exposes the ListPaymentRules use case as a built-in self tool.
type ListPaymentRulesTool struct {
	listPaymentRulesExecutor tool_interfaces.ListPaymentRulesExecutorInterface
}

func NewListPaymentRulesTool(listPaymentRulesExecutor tool_interfaces.ListPaymentRulesExecutorInterface) *ListPaymentRulesTool {
	return &ListPaymentRulesTool{listPaymentRulesExecutor: listPaymentRulesExecutor}
}

func (t *ListPaymentRulesTool) Definition() *models.Tool {
	return &models.Tool{
		Name:        "list_payment_rules",
		Description: "List your payment rules (pricing for callers).",
		InputScheme: utils.EmptyObjectScheme(),
	}
}

func (t *ListPaymentRulesTool) Execute(ctx context.Context, _ map[string]any) (map[string]any, error) {
	return utils.ToToolResult(t.listPaymentRulesExecutor.Execute(ctx))
}

func (t *ListPaymentRulesTool) IsEnabledByDefault() bool {
	return false
}
