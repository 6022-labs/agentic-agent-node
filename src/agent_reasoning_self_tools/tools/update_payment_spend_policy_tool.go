package tools

import (
	"context"

	payment_requests "github.com/6022-labs/agentic-agent-node/src/agent_payment/use_cases/requests"
	"github.com/6022-labs/agentic-agent-node/src/agent_reasoning/models"
	tool_interfaces "github.com/6022-labs/agentic-agent-node/src/agent_reasoning_self_tools/tools/interfaces"
	"github.com/6022-labs/agentic-agent-node/src/agent_reasoning_self_tools/utils"
)

// UpdatePaymentSpendPolicyTool is off by default: enabling it lets the agent choose whom it pays.
type UpdatePaymentSpendPolicyTool struct {
	updatePaymentSpendPolicyExecutor tool_interfaces.UpdatePaymentSpendPolicyExecutorInterface
}

func NewUpdatePaymentSpendPolicyTool(
	updatePaymentSpendPolicyExecutor tool_interfaces.UpdatePaymentSpendPolicyExecutorInterface,
) *UpdatePaymentSpendPolicyTool {
	return &UpdatePaymentSpendPolicyTool{updatePaymentSpendPolicyExecutor: updatePaymentSpendPolicyExecutor}
}

func (t *UpdatePaymentSpendPolicyTool) Definition() *models.Tool {
	return &models.Tool{
		Name:        "update_payment_spend_policy",
		Description: "Replace the list of payees you may pay, whatever the token; an empty list pays nobody.",
		InputScheme: utils.ObjectScheme(
			map[string]models.ToolInputSchemeProperty{
				"payToPatterns": utils.ArrayProperty(
					"Payee addresses in any case, or regexes written in lowercase (\".*\" allows every payee).",
					utils.StringProperty("One payee regex."),
				),
			},
			[]string{"payToPatterns"},
		),
	}
}

func (t *UpdatePaymentSpendPolicyTool) Execute(ctx context.Context, args map[string]any) (map[string]any, error) {
	payToPatterns, err := utils.JsonArrayArg[string](args, "payToPatterns")
	if err != nil {
		return nil, err
	}
	request := &payment_requests.UpdatePaymentSpendPolicyRequest{PayToPatterns: payToPatterns}
	return utils.ToToolResult(t.updatePaymentSpendPolicyExecutor.Execute(ctx, request))
}

func (t *UpdatePaymentSpendPolicyTool) IsEnabledByDefault() bool {
	return false
}

func (t *UpdatePaymentSpendPolicyTool) GroupName() string {
	return GroupPayments
}
