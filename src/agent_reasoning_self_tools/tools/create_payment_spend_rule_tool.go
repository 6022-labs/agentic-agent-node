package tools

import (
	"context"

	"github.com/6022-labs/agentic-agent-node/src/agent_reasoning/models"
	tool_interfaces "github.com/6022-labs/agentic-agent-node/src/agent_reasoning_self_tools/tools/interfaces"
	"github.com/6022-labs/agentic-agent-node/src/agent_reasoning_self_tools/utils"
)

// CreatePaymentSpendRuleTool is off by default: enabling it lets the agent widen its own spending.
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
		InputScheme: utils.ObjectScheme(
			t.savePaymentSpendRuleProperties(),
			t.savePaymentSpendRuleRequired(),
		),
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
