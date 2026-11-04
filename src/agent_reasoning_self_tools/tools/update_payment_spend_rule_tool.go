package tools

import (
	"context"

	payment_requests "github.com/6022-labs/agentic-agent-node/src/agent_payment/use_cases/requests"
	"github.com/6022-labs/agentic-agent-node/src/agent_reasoning/models"
	tool_interfaces "github.com/6022-labs/agentic-agent-node/src/agent_reasoning_self_tools/tools/interfaces"
	"github.com/6022-labs/agentic-agent-node/src/agent_reasoning_self_tools/utils"
)

// UpdatePaymentSpendRuleTool is off by default: enabling it lets the agent raise its own caps.
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
		Name: "update_payment_spend_rule",
		Description: "Replace one of your spend rules by id. Omitted fields are cleared: resend the cap and " +
			"every window to keep them. A window keeps its spend only if its windowSeconds and the token stay.",
		InputScheme: utils.ObjectScheme(
			t.propertiesWithId(),
			append([]string{"id"}, t.savePaymentSpendRuleRequired()...),
		),
	}
}

// propertiesWithId adds the route id the update tool needs on top of the shared rule scheme.
func (t *UpdatePaymentSpendRuleTool) propertiesWithId() map[string]models.ToolInputSchemeProperty {
	properties := t.savePaymentSpendRuleProperties()
	properties["id"] = utils.StringProperty("The rule id.")
	return properties
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
