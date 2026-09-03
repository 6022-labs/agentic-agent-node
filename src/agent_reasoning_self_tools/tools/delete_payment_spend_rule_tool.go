package tools

import (
	"context"

	payment_requests "github.com/6022-labs/agentic-agent-node/src/agent_payment/use_cases/requests"
	"github.com/6022-labs/agentic-agent-node/src/agent_reasoning/models"
	tool_interfaces "github.com/6022-labs/agentic-agent-node/src/agent_reasoning_self_tools/tools/interfaces"
	"github.com/6022-labs/agentic-agent-node/src/agent_reasoning_self_tools/utils"
)

// DeletePaymentSpendRuleTool exposes the DeletePaymentSpendRule use case as a built-in self tool.
type DeletePaymentSpendRuleTool struct {
	deletePaymentSpendRuleExecutor tool_interfaces.DeletePaymentSpendRuleExecutorInterface
}

func NewDeletePaymentSpendRuleTool(
	deletePaymentSpendRuleExecutor tool_interfaces.DeletePaymentSpendRuleExecutorInterface,
) *DeletePaymentSpendRuleTool {
	return &DeletePaymentSpendRuleTool{deletePaymentSpendRuleExecutor: deletePaymentSpendRuleExecutor}
}

func (t *DeletePaymentSpendRuleTool) Definition() *models.Tool {
	return &models.Tool{
		Name:        "delete_payment_spend_rule",
		Description: "Remove one of your spend rules by id. Quotes only that rule allowed will be refused.",
		InputScheme: utils.ObjectScheme(map[string]models.ToolInputSchemeProperty{
			"id": utils.StringProperty("The rule id."),
		}, []string{"id"}),
	}
}

func (t *DeletePaymentSpendRuleTool) Execute(ctx context.Context, args map[string]any) (map[string]any, error) {
	id, err := utils.UuidArg(args, "id")
	if err != nil {
		return nil, err
	}
	if err := t.deletePaymentSpendRuleExecutor.Execute(
		ctx,
		&payment_requests.DeletePaymentSpendRuleRequest{Id: id},
	); err != nil {
		return nil, err
	}
	return utils.SuccessResult(), nil
}

func (t *DeletePaymentSpendRuleTool) IsEnabledByDefault() bool {
	return false
}

func (t *DeletePaymentSpendRuleTool) GroupName() string {
	return GroupPayments
}
