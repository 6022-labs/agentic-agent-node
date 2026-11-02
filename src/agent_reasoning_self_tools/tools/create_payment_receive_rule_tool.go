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
		Name: "create_payment_receive_rule",
		Description: "Create the pricing rule of one token for your callers; a token already priced " +
			"must be updated instead. Config names the token (x402 by default).",
		InputScheme: utils.ObjectScheme(
			t.savePaymentReceiveRuleProperties(),
			t.savePaymentReceiveRuleRequired(),
		),
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
