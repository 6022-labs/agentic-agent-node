package tools

import (
	"context"

	payment_requests "github.com/6022-labs/agentic-agent-node/src/agent_payment/use_cases/requests"
	"github.com/6022-labs/agentic-agent-node/src/agent_reasoning/models"
	tool_interfaces "github.com/6022-labs/agentic-agent-node/src/agent_reasoning_self_tools/tools/interfaces"
	"github.com/6022-labs/agentic-agent-node/src/agent_reasoning_self_tools/utils"
)

// DeletePaymentReceiveRuleTool exposes the DeletePaymentReceiveRule use case as a built-in self tool.
type DeletePaymentReceiveRuleTool struct {
	deletePaymentReceiveRuleExecutor tool_interfaces.DeletePaymentReceiveRuleExecutorInterface
}

func NewDeletePaymentReceiveRuleTool(
	deletePaymentReceiveRuleExecutor tool_interfaces.DeletePaymentReceiveRuleExecutorInterface,
) *DeletePaymentReceiveRuleTool {
	return &DeletePaymentReceiveRuleTool{deletePaymentReceiveRuleExecutor: deletePaymentReceiveRuleExecutor}
}

func (t *DeletePaymentReceiveRuleTool) Definition() *models.Tool {
	return &models.Tool{
		Name:        "delete_payment_receive_rule",
		Description: "Delete one of your pricing rules.",
		InputScheme: utils.ObjectScheme(map[string]models.ToolInputSchemeProperty{
			"id": utils.StringProperty("Id of the rule to delete (see list_payment_receive_rules)."),
		}, []string{"id"}),
	}
}

func (t *DeletePaymentReceiveRuleTool) Execute(ctx context.Context, args map[string]any) (map[string]any, error) {
	id, err := utils.UuidArg(args, "id")
	if err != nil {
		return nil, err
	}
	if err := t.deletePaymentReceiveRuleExecutor.Execute(
		ctx,
		&payment_requests.DeletePaymentReceiveRuleRequest{Id: id},
	); err != nil {
		return nil, err
	}
	return utils.SuccessResult(), nil
}

func (t *DeletePaymentReceiveRuleTool) IsEnabledByDefault() bool {
	return false
}

func (t *DeletePaymentReceiveRuleTool) GroupName() string {
	return GroupPayments
}
