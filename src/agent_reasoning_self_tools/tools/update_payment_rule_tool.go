package tools

import (
	"context"

	payment_requests "github.com/6022-labs/agentic-agent-node/src/agent_payment/use_cases/requests"
	"github.com/6022-labs/agentic-agent-node/src/agent_reasoning/models"
	tool_interfaces "github.com/6022-labs/agentic-agent-node/src/agent_reasoning_self_tools/tools/interfaces"
	"github.com/6022-labs/agentic-agent-node/src/agent_reasoning_self_tools/utils"
)

// UpdatePaymentRuleTool exposes the UpdatePaymentRule use case as a built-in self tool.
type UpdatePaymentRuleTool struct {
	*AbstractPaymentRuleTool
	updatePaymentRuleExecutor tool_interfaces.UpdatePaymentRuleExecutorInterface
}

func NewUpdatePaymentRuleTool(
	updatePaymentRuleExecutor tool_interfaces.UpdatePaymentRuleExecutorInterface,
) *UpdatePaymentRuleTool {
	return &UpdatePaymentRuleTool{
		AbstractPaymentRuleTool:   NewAbstractPaymentRuleTool(),
		updatePaymentRuleExecutor: updatePaymentRuleExecutor,
	}
}

func (t *UpdatePaymentRuleTool) Definition() *models.Tool {
	return &models.Tool{
		Name:        "update_payment_rule",
		Description: "Replace one of your pricing rules (identity and position preserved).",
		InputScheme: utils.ObjectScheme(map[string]models.ToolInputSchemeProperty{
			"id":     utils.StringProperty("Id of the rule to update (see list_payment_rules)."),
			"name":   utils.StringProperty("Rule name (1-64 characters)."),
			"price":  utils.StringProperty("Non-negative integer price in the asset's atomic units (0 = free)."),
			"method": utils.StringProperty("Payment method; defaults to x402."),
			"config": utils.ObjectProperty(
				"Method-specific rule configuration. For x402: {network (CAIP-2), asset (ERC-20 address), payTo (address), assetTransferMethod, addressPatterns}.",
			),
		}, []string{"id", "name", "price", "config"}),
	}
}

func (t *UpdatePaymentRuleTool) Execute(ctx context.Context, args map[string]any) (map[string]any, error) {
	id, err := utils.UuidArg(args, "id")
	if err != nil {
		return nil, err
	}
	saveRequest, err := t.savePaymentRuleRequestFromArgs(args)
	if err != nil {
		return nil, err
	}
	request := &payment_requests.UpdatePaymentRuleRequest{
		Id:                     id,
		SavePaymentRuleRequest: *saveRequest,
	}

	return utils.ToToolResult(t.updatePaymentRuleExecutor.Execute(ctx, request))
}

func (t *UpdatePaymentRuleTool) IsEnabledByDefault() bool {
	return false
}
