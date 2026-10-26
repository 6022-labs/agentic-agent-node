package tools

import (
	payment_requests "github.com/6022-labs/agentic-agent-node/src/agent_payment/use_cases/requests"
	"github.com/6022-labs/agentic-agent-node/src/agent_reasoning_self_tools/utils"
)

// AbstractPaymentReceiveRuleTool shares the args → SavePaymentReceiveRuleRequest mapping
// between the create/update payment-receive-rule tools.
type AbstractPaymentReceiveRuleTool struct{}

func NewAbstractPaymentReceiveRuleTool() *AbstractPaymentReceiveRuleTool {
	return &AbstractPaymentReceiveRuleTool{}
}

func (t *AbstractPaymentReceiveRuleTool) savePaymentReceiveRuleRequestFromArgs(
	args map[string]any,
) (*payment_requests.SavePaymentReceiveRuleRequest, error) {
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
	return &payment_requests.SavePaymentReceiveRuleRequest{
		Name:   name,
		Price:  price,
		Method: utils.OptionalStringArg(args, "method"),
		Config: config,
	}, nil
}
