package tools

import (
	payment_requests "github.com/6022-labs/agentic-agent-node/src/agent_payment/use_cases/requests"
	"github.com/6022-labs/agentic-agent-node/src/agent_reasoning_self_tools/utils"
)

// AbstractPaymentSpendRuleTool shares the args → SavePaymentSpendRuleRequest mapping
// between the create/update payment-spend-rule tools.
type AbstractPaymentSpendRuleTool struct{}

func NewAbstractPaymentSpendRuleTool() *AbstractPaymentSpendRuleTool {
	return &AbstractPaymentSpendRuleTool{}
}

func (t *AbstractPaymentSpendRuleTool) savePaymentSpendRuleRequestFromArgs(
	args map[string]any,
) (*payment_requests.SavePaymentSpendRuleRequest, error) {
	name, err := utils.StringArg(args, "name")
	if err != nil {
		return nil, err
	}
	maxAmount, err := utils.StringArg(args, "maxAmount")
	if err != nil {
		return nil, err
	}
	config, err := utils.RawJsonObjectArg(args, "config")
	if err != nil {
		return nil, err
	}
	return &payment_requests.SavePaymentSpendRuleRequest{
		Name:      name,
		MaxAmount: maxAmount,
		Method:    utils.OptionalStringArg(args, "method"),
		Config:    config,
	}, nil
}
