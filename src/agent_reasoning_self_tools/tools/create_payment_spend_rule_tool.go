package tools

import (
	"context"

	"github.com/6022-labs/agentic-agent-node/src/agent_reasoning/models"
	tool_interfaces "github.com/6022-labs/agentic-agent-node/src/agent_reasoning_self_tools/tools/interfaces"
	"github.com/6022-labs/agentic-agent-node/src/agent_reasoning_self_tools/utils"
)

// CreatePaymentSpendRuleTool exposes the CreatePaymentSpendRule use case as a built-in self tool.
// Enabling it lets the agent widen its own spending cap, so it is off unless an owner turns it on.
type CreatePaymentSpendRuleTool struct {
	*AbstractPaymentSpendRuleTool
	createPaymentSpendRuleExecutor tool_interfaces.CreatePaymentSpendRuleExecutorInterface
}

func NewCreatePaymentSpendRuleTool(
	createPaymentSpendRuleExecutor tool_interfaces.CreatePaymentSpendRuleExecutorInterface,
) *CreatePaymentSpendRuleTool {
	return &CreatePaymentSpendRuleTool{
		AbstractPaymentSpendRuleTool:   NewAbstractPaymentSpendRuleTool(),
		createPaymentSpendRuleExecutor: createPaymentSpendRuleExecutor,
	}
}

func (t *CreatePaymentSpendRuleTool) Definition() *models.Tool {
	return &models.Tool{
		Name: "create_payment_spend_rule",
		Description: "Create a rule allowing yourself to pay. Every facet of config must match a quote for it " +
			"to be authorized, and the smallest matching maxAmountPerPayment is the cap.",
		InputScheme: utils.ObjectScheme(map[string]models.ToolInputSchemeProperty{
			"name": utils.StringProperty("Rule name (1-64 characters)."),
			"maxAmountPerPayment": utils.StringProperty(
				"Non-negative integer cap in the asset's atomic units (0 denies anything priced).",
			),
			"method": utils.StringProperty("Payment method; defaults to x402."),
			"config": utils.ObjectProperty(
				"Method-specific matcher. For x402: {networkPatterns, assetPatterns, payToPatterns, " +
					"assetTransferMethodPatterns}, each a non-empty array of regexes (\".*\" matches everything).",
			),
		}, []string{"name", "maxAmountPerPayment", "config"}),
	}
}

func (t *CreatePaymentSpendRuleTool) Execute(ctx context.Context, args map[string]any) (map[string]any, error) {
	request, err := t.savePaymentSpendRuleRequestFromArgs(args)
	if err != nil {
		return nil, err
	}
	return utils.ToToolResult(t.createPaymentSpendRuleExecutor.Execute(ctx, request))
}

func (t *CreatePaymentSpendRuleTool) IsEnabledByDefault() bool {
	return false
}

func (t *CreatePaymentSpendRuleTool) GroupName() string {
	return GroupPayments
}
