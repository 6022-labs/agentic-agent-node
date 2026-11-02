package tools

import (
	"context"

	payment_requests "github.com/6022-labs/agentic-agent-node/src/agent_payment/use_cases/requests"
	"github.com/6022-labs/agentic-agent-node/src/agent_reasoning/models"
	tool_interfaces "github.com/6022-labs/agentic-agent-node/src/agent_reasoning_self_tools/tools/interfaces"
	"github.com/6022-labs/agentic-agent-node/src/agent_reasoning_self_tools/utils"
)

// UpdatePaymentReceiveRuleTool exposes the UpdatePaymentReceiveRule use case as a built-in self tool.
type UpdatePaymentReceiveRuleTool struct {
	*AbstractPaymentReceiveRuleTool
	updatePaymentReceiveRuleExecutor tool_interfaces.UpdatePaymentReceiveRuleExecutorInterface
}

func NewUpdatePaymentReceiveRuleTool(
	updatePaymentReceiveRuleExecutor tool_interfaces.UpdatePaymentReceiveRuleExecutorInterface,
) *UpdatePaymentReceiveRuleTool {
	return &UpdatePaymentReceiveRuleTool{
		AbstractPaymentReceiveRuleTool:   NewAbstractPaymentReceiveRuleTool(),
		updatePaymentReceiveRuleExecutor: updatePaymentReceiveRuleExecutor,
	}
}

func (t *UpdatePaymentReceiveRuleTool) Definition() *models.Tool {
	return &models.Tool{
		Name:        "update_payment_receive_rule",
		Description: "Replace one of your pricing rules: its name, token and every price.",
		InputScheme: utils.ObjectScheme(
			t.propertiesWithId(),
			append([]string{"id"}, t.savePaymentReceiveRuleRequired()...),
		),
	}
}

// propertiesWithId adds the route id the update tool needs on top of the shared rule scheme.
func (t *UpdatePaymentReceiveRuleTool) propertiesWithId() map[string]models.ToolInputSchemeProperty {
	properties := t.savePaymentReceiveRuleProperties()
	properties["id"] = utils.StringProperty("Id of the rule to update (see list_payment_receive_rules).")
	return properties
}

func (t *UpdatePaymentReceiveRuleTool) Execute(ctx context.Context, args map[string]any) (map[string]any, error) {
	id, err := utils.UuidArg(args, "id")
	if err != nil {
		return nil, err
	}
	saveRequest, err := t.savePaymentReceiveRuleRequestFromArgs(args)
	if err != nil {
		return nil, err
	}
	request := &payment_requests.UpdatePaymentReceiveRuleRequest{
		Id:                            id,
		SavePaymentReceiveRuleRequest: *saveRequest,
	}

	return utils.ToToolResult(t.updatePaymentReceiveRuleExecutor.Execute(ctx, request))
}

func (t *UpdatePaymentReceiveRuleTool) IsEnabledByDefault() bool {
	return false
}

func (t *UpdatePaymentReceiveRuleTool) GroupName() string {
	return GroupPayments
}
