package tools

import (
	payment_requests "github.com/6022-labs/agentic-agent-node/src/agent_payment/use_cases/requests"
	"github.com/6022-labs/agentic-agent-node/src/agent_reasoning_self_tools/utils"
)

// AbstractPaymentRuleTool shares the args → SavePaymentRuleRequest mapping
// between the create/update payment-rule tools.
type AbstractPaymentRuleTool struct{}

func NewAbstractPaymentRuleTool() *AbstractPaymentRuleTool {
	return &AbstractPaymentRuleTool{}
}

func (t *AbstractPaymentRuleTool) savePaymentRuleRequestFromArgs(
	args map[string]any,
) (*payment_requests.SavePaymentRuleRequest, error) {
	name, err := utils.StringArg(args, "name")
	if err != nil {
		return nil, err
	}
	price, err := utils.StringArg(args, "price")
	if err != nil {
		return nil, err
	}
	config, err := utils.RawJsonObjectArg(args, "config")
	if err != nil {
		return nil, err
	}
	return &payment_requests.SavePaymentRuleRequest{
		Name:   name,
		Price:  price,
		Method: utils.OptionalStringArg(args, "method"),
		Config: config,
	}, nil
}
