package tools

import (
	"context"
	"github.com/6022-labs/agentic-agent-node/src/agent_reasoning/models"
	tool_interfaces "github.com/6022-labs/agentic-agent-node/src/agent_reasoning_self_tools/tools/interfaces"
	"github.com/6022-labs/agentic-agent-node/src/agent_reasoning_self_tools/utils"
)

// DeletePaymentRuleTool exposes the DeletePaymentRule use case as a built-in self tool.
type DeletePaymentRuleTool struct {
	deletePaymentRuleExecutor tool_interfaces.DeletePaymentRuleExecutorInterface
}

func NewDeletePaymentRuleTool(deletePaymentRuleExecutor tool_interfaces.DeletePaymentRuleExecutorInterface) *DeletePaymentRuleTool {
	return &DeletePaymentRuleTool{deletePaymentRuleExecutor: deletePaymentRuleExecutor}
}

func (t *DeletePaymentRuleTool) Definition() *models.Tool {
	return &models.Tool{
		Name:        "delete_payment_rule",
		Description: "Delete one of your pricing rules.",
		InputScheme: utils.ObjectScheme(map[string]models.ToolInputSchemeProperty{
			"id": utils.StringProperty("Id of the rule to delete (see list_payment_rules)."),
		}, []string{"id"}),
	}
}

func (t *DeletePaymentRuleTool) Execute(ctx context.Context, args map[string]any) (map[string]any, error) {
	id, err := utils.UuidArg(args, "id")
	if err != nil {
		return nil, err
	}
	if err := t.deletePaymentRuleExecutor.Execute(ctx, id); err != nil {
		return nil, err
	}
	return utils.SuccessResult(), nil
}
