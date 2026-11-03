package tools

import (
	payment_requests "github.com/6022-labs/agentic-agent-node/src/agent_payment/use_cases/requests"
	"github.com/6022-labs/agentic-agent-node/src/agent_reasoning/models"
	"github.com/6022-labs/agentic-agent-node/src/agent_reasoning_self_tools/utils"
)

// AbstractPaymentReceiveRuleTool shares the rule scheme and args mapping between the create and update tools.
type AbstractPaymentReceiveRuleTool struct{}

func NewAbstractPaymentReceiveRuleTool() *AbstractPaymentReceiveRuleTool {
	return &AbstractPaymentReceiveRuleTool{}
}

func (t *AbstractPaymentReceiveRuleTool) savePaymentReceiveRuleProperties() map[string]models.ToolInputSchemeProperty {
	price := utils.ObjectProperty(
		"{price (non-negative integer string in the asset's atomic units, 0 = free), " +
			"addressPatterns (payer addresses in any case, or regexes written in lowercase; \".*\" = anyone)}.",
	)
	return map[string]models.ToolInputSchemeProperty{
		"name": utils.StringProperty("Rule name (1-64 characters)."),
		"prices": utils.ArrayProperty(
			"At least one price, each amount at most once; a payer pays the cheapest price whose patterns "+
				"match them, and only prices open to \".*\" are advertised in a 402.",
			price,
		),
		"method": utils.StringProperty("Payment method; defaults to x402."),
		"config": utils.ObjectProperty(
			"Method-specific token. For x402: {network (CAIP-2, e.g. eip155:8453), asset (ERC-20 address), " +
				"payTo (address receiving the payment)}; the transfer method is read from the token. " +
				"One rule per token.",
		),
	}
}

func (t *AbstractPaymentReceiveRuleTool) savePaymentReceiveRuleRequired() []string {
	return []string{"name", "prices", "config"}
}

func (t *AbstractPaymentReceiveRuleTool) savePaymentReceiveRuleRequestFromArgs(
	args map[string]any,
) (*payment_requests.SavePaymentReceiveRuleRequest, error) {
	name, err := utils.StringArg(args, "name")
	if err != nil {
		return nil, err
	}
	prices, err := utils.JsonArrayArg[payment_requests.SavePaymentReceiveRulePriceRequest](args, "prices")
	if err != nil {
		return nil, err
	}
	config, err := utils.RawJsonObjectArg(args, "config")
	if err != nil {
		return nil, err
	}
	return &payment_requests.SavePaymentReceiveRuleRequest{
		SavePaymentRuleRequest: payment_requests.SavePaymentRuleRequest{
			Name:   name,
			Method: utils.OptionalStringArg(args, "method"),
		},
		Prices: prices,
		Config: config,
	}, nil
}
