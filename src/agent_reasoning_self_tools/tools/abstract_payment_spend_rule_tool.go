package tools

import (
	payment_requests "github.com/6022-labs/agentic-agent-node/src/agent_payment/use_cases/requests"
	"github.com/6022-labs/agentic-agent-node/src/agent_reasoning/models"
	"github.com/6022-labs/agentic-agent-node/src/agent_reasoning_self_tools/utils"
	"github.com/samber/lo"
)

// AbstractPaymentSpendRuleTool shares the rule scheme and its args → SavePaymentSpendRuleRequest
// mapping between the create/update payment-spend-rule tools, so the two cannot drift apart.
type AbstractPaymentSpendRuleTool struct{}

func NewAbstractPaymentSpendRuleTool() *AbstractPaymentSpendRuleTool {
	return &AbstractPaymentSpendRuleTool{}
}

func (t *AbstractPaymentSpendRuleTool) savePaymentSpendRuleProperties() map[string]models.ToolInputSchemeProperty {
	return map[string]models.ToolInputSchemeProperty{
		"name": utils.StringProperty("Rule name (1-64 characters)."),
		"maxAmountPerPayment": utils.StringProperty(
			"Non-negative integer cap in the asset's atomic units (0 denies anything priced).",
		),
		"windowSeconds": utils.StringProperty(
			"Optional recurring period in seconds (86400 = a day); omit for a per-payment cap only.",
		),
		"maxAmountPerWindow": utils.StringProperty(
			"Budget for one period, in the same atomic units; required when windowSeconds is set, " +
				"and never below maxAmountPerPayment.",
		),
		"method": utils.StringProperty("Payment method; defaults to x402."),
		"config": utils.ObjectProperty(
			"Method-specific matcher. For x402: {networkPatterns, assetPatterns, payToPatterns, " +
				"assetTransferMethodPatterns}, each a non-empty array of regexes (\".*\" matches everything).",
		),
	}
}

func (t *AbstractPaymentSpendRuleTool) savePaymentSpendRuleRequired() []string {
	return []string{"name", "maxAmountPerPayment", "config"}
}

func (t *AbstractPaymentSpendRuleTool) savePaymentSpendRuleRequestFromArgs(
	args map[string]any,
) (*payment_requests.SavePaymentSpendRuleRequest, error) {
	name, err := utils.StringArg(args, "name")
	if err != nil {
		return nil, err
	}
	maxAmountPerPayment, err := utils.StringArg(args, "maxAmountPerPayment")
	if err != nil {
		return nil, err
	}
	config, err := utils.RawJsonObjectArg(args, "config")
	if err != nil {
		return nil, err
	}
	return &payment_requests.SavePaymentSpendRuleRequest{
		Name:                name,
		MaxAmountPerPayment: maxAmountPerPayment,
		WindowSeconds:       lo.FromPtr(utils.OptionalIntArg(args, "windowSeconds")),
		MaxAmountPerWindow:  utils.OptionalStringArg(args, "maxAmountPerWindow"),
		Method:              utils.OptionalStringArg(args, "method"),
		Config:              config,
	}, nil
}
