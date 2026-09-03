package tools

import (
	"context"

	payment_requests "github.com/6022-labs/agentic-agent-node/src/agent_payment/use_cases/requests"
	"github.com/6022-labs/agentic-agent-node/src/agent_reasoning/models"
	tool_interfaces "github.com/6022-labs/agentic-agent-node/src/agent_reasoning_self_tools/tools/interfaces"
	"github.com/6022-labs/agentic-agent-node/src/agent_reasoning_self_tools/utils"
)

// UpdatePaymentSpendRuleTool exposes the UpdatePaymentSpendRule use case as a built-in self tool.
// Enabling it lets the agent raise its own cap, so it is off unless an owner turns it on.
type UpdatePaymentSpendRuleTool struct {
	*AbstractPaymentSpendRuleTool
	updatePaymentSpendRuleExecutor tool_interfaces.UpdatePaymentSpendRuleExecutorInterface
}

func NewUpdatePaymentSpendRuleTool(
	updatePaymentSpendRuleExecutor tool_interfaces.UpdatePaymentSpendRuleExecutorInterface,
) *UpdatePaymentSpendRuleTool {
	return &UpdatePaymentSpendRuleTool{
		AbstractPaymentSpendRuleTool:   NewAbstractPaymentSpendRuleTool(),
		updatePaymentSpendRuleExecutor: updatePaymentSpendRuleExecutor,
	}
}

func (t *UpdatePaymentSpendRuleTool) Definition() *models.Tool {
	return &models.Tool{
		Name:        "update_payment_spend_rule",
		Description: "Replace one of your spend rules by id.",
		InputScheme: utils.ObjectScheme(map[string]models.ToolInputSchemeProperty{
			"id":   utils.StringProperty("The rule id."),
			"name": utils.StringProperty("Rule name (1-64 characters)."),
			"maxAmount": utils.StringProperty(
				"Non-negative integer cap in the asset's atomic units (0 denies anything priced).",
			),
			"method": utils.StringProperty("Payment method; defaults to x402."),
			"config": utils.ObjectProperty(
				"Method-specific matcher. For x402: {networkPatterns, assetPatterns, payToPatterns, " +
					"assetTransferMethodPatterns}, each a non-empty array of regexes.",
			),
		}, []string{"id", "name", "maxAmount", "config"}),
	}
}

func (t *UpdatePaymentSpendRuleTool) Execute(ctx context.Context, args map[string]any) (map[string]any, error) {
	id, err := utils.UuidArg(args, "id")
	if err != nil {
		return nil, err
	}
	request, err := t.savePaymentSpendRuleRequestFromArgs(args)
	if err != nil {
		return nil, err
	}
	return utils.ToToolResult(
		t.updatePaymentSpendRuleExecutor.Execute(ctx, &payment_requests.UpdatePaymentSpendRuleRequest{
			Id:                          id,
			SavePaymentSpendRuleRequest: *request,
		}),
	)
}

func (t *UpdatePaymentSpendRuleTool) IsEnabledByDefault() bool {
	return false
}

func (t *UpdatePaymentSpendRuleTool) GroupName() string {
	return GroupPayments
}
