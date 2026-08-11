package tools

import (
	"context"

	"github.com/6022-labs/agentic-agent-node/src/agent_reasoning/models"
	"github.com/6022-labs/agentic-agent-node/src/agent_reasoning_self_tools/utils"

	tool_interfaces "github.com/6022-labs/agentic-agent-node/src/agent_reasoning_self_tools/tools/interfaces"
)

// CreatePaymentRuleTool exposes the CreatePaymentRule use case as a built-in self tool.
type CreatePaymentRuleTool struct {
	*AbstractPaymentRuleTool
	createPaymentRuleExecutor tool_interfaces.CreatePaymentRuleExecutorInterface
}

func NewCreatePaymentRuleTool(
	createPaymentRuleExecutor tool_interfaces.CreatePaymentRuleExecutorInterface,
) *CreatePaymentRuleTool {
	return &CreatePaymentRuleTool{
		AbstractPaymentRuleTool:   NewAbstractPaymentRuleTool(),
		createPaymentRuleExecutor: createPaymentRuleExecutor,
	}
}

func (t *CreatePaymentRuleTool) Definition() *models.Tool {
	return &models.Tool{
		Name:        "create_payment_rule",
		Description: "Create a pricing rule for your callers. Config carries the method-specific settings (x402 by default).",
		InputScheme: utils.ObjectScheme(map[string]models.ToolInputSchemeProperty{
			"name":   utils.StringProperty("Rule name (1-64 characters)."),
			"price":  utils.StringProperty("Non-negative integer price in the asset's atomic units (0 = free)."),
			"method": utils.StringProperty("Payment method; defaults to x402."),
			"config": utils.ObjectProperty(
				"Method-specific rule configuration. For x402: {network (CAIP-2), asset (ERC-20 address), payTo (address), assetTransferMethod, addressPatterns}.",
			),
		}, []string{"name", "price", "config"}),
	}
}

func (t *CreatePaymentRuleTool) Execute(ctx context.Context, args map[string]any) (map[string]any, error) {
	request, err := t.savePaymentRuleRequestFromArgs(args)
	if err != nil {
		return nil, err
	}
	return utils.ToToolResult(t.createPaymentRuleExecutor.Execute(ctx, request))
}

func (t *CreatePaymentRuleTool) IsEnabledByDefault() bool {
	return false
}

func (t *CreatePaymentRuleTool) GroupName() string {
	return GroupPayments
}
