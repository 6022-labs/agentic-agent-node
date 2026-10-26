package tools

import (
	"context"

	"github.com/6022-labs/agentic-agent-node/src/agent_reasoning/models"
	"github.com/6022-labs/agentic-agent-node/src/agent_reasoning_self_tools/utils"

	tool_interfaces "github.com/6022-labs/agentic-agent-node/src/agent_reasoning_self_tools/tools/interfaces"
)

// CreatePaymentReceiveRuleTool exposes the CreatePaymentReceiveRule use case as a built-in self tool.
type CreatePaymentReceiveRuleTool struct {
	*AbstractPaymentReceiveRuleTool
	createPaymentReceiveRuleExecutor tool_interfaces.CreatePaymentReceiveRuleExecutorInterface
}

func NewCreatePaymentReceiveRuleTool(
	createPaymentReceiveRuleExecutor tool_interfaces.CreatePaymentReceiveRuleExecutorInterface,
) *CreatePaymentReceiveRuleTool {
	return &CreatePaymentReceiveRuleTool{
		AbstractPaymentReceiveRuleTool:   NewAbstractPaymentReceiveRuleTool(),
		createPaymentReceiveRuleExecutor: createPaymentReceiveRuleExecutor,
	}
}

func (t *CreatePaymentReceiveRuleTool) Definition() *models.Tool {
	return &models.Tool{
		Name:        "create_payment_receive_rule",
		Description: "Create a pricing rule for your callers. Config carries the method-specific settings (x402 by default).",
		InputScheme: utils.ObjectScheme(map[string]models.ToolInputSchemeProperty{
			"name":   utils.StringProperty("Rule name (1-64 characters)."),
			"price":  utils.StringProperty("Non-negative integer price in the asset's atomic units (0 = free)."),
			"method": utils.StringProperty("Payment method; defaults to x402."),
			"config": utils.ObjectProperty(
				"Method-specific rule configuration. For x402: {network (CAIP-2), asset (ERC-20 address), " +
					"payTo (address), assetTransferMethod, addressPatterns}.",
			),
		}, []string{"name", "price", "config"}),
	}
}

func (t *CreatePaymentReceiveRuleTool) Execute(ctx context.Context, args map[string]any) (map[string]any, error) {
	request, err := t.savePaymentReceiveRuleRequestFromArgs(args)
	if err != nil {
		return nil, err
	}
	return utils.ToToolResult(t.createPaymentReceiveRuleExecutor.Execute(ctx, request))
}

func (t *CreatePaymentReceiveRuleTool) IsEnabledByDefault() bool {
	return false
}

func (t *CreatePaymentReceiveRuleTool) GroupName() string {
	return GroupPayments
}
