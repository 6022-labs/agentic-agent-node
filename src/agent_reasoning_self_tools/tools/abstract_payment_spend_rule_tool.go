package tools

import (
	payment_requests "github.com/6022-labs/agentic-agent-node/src/agent_payment/use_cases/requests"
	"github.com/6022-labs/agentic-agent-node/src/agent_reasoning/models"
	"github.com/6022-labs/agentic-agent-node/src/agent_reasoning_self_tools/utils"
	"github.com/samber/lo"
)

// AbstractPaymentSpendRuleTool shares the rule scheme and args mapping between the create and update tools.
type AbstractPaymentSpendRuleTool struct{}

func NewAbstractPaymentSpendRuleTool() *AbstractPaymentSpendRuleTool {
	return &AbstractPaymentSpendRuleTool{}
}

func (t *AbstractPaymentSpendRuleTool) savePaymentSpendRuleProperties() map[string]models.ToolInputSchemeProperty {
	window := utils.ObjectProperty(
		"{windowSeconds (integer, 86400 = a day), maxAmount (non-negative integer string in atomic units)}.",
	)
	return map[string]models.ToolInputSchemeProperty{
		"name": utils.StringProperty("Rule name (1-64 characters)."),
		"maxAmountPerPayment": utils.StringProperty(
			"Optional cap on one payment in the asset's atomic units (0 denies anything priced).",
		),
		"windows": utils.ArrayProperty(
			"Recurring budgets, at most one per windowSeconds; the rule needs a cap, a window, or both.",
			window,
		),
		"method": utils.StringProperty("Payment method; defaults to x402."),
		"config": utils.ObjectProperty(
			"Method-specific token. For x402: {network (CAIP-2, e.g. eip155:8453), asset (ERC-20 address), " +
				"assetTransferMethod (optional: eip3009 or permit2; omit for either)}. One rule per token.",
		),
	}
}

func (t *AbstractPaymentSpendRuleTool) savePaymentSpendRuleRequired() []string {
	return []string{"name", "config"}
}

func (t *AbstractPaymentSpendRuleTool) savePaymentSpendRuleRequestFromArgs(
	args map[string]any,
) (*payment_requests.SavePaymentSpendRuleRequest, error) {
	name, err := utils.StringArg(args, "name")
	if err != nil {
		return nil, err
	}
	windows, err := utils.OptionalJsonArrayArg[payment_requests.SavePaymentSpendRuleWindowRequest](args, "windows")
	if err != nil {
		return nil, err
	}
	config, err := utils.RawJsonObjectArg(args, "config")
	if err != nil {
		return nil, err
	}
	return &payment_requests.SavePaymentSpendRuleRequest{
		Name:                name,
		MaxAmountPerPayment: lo.EmptyableToPtr(utils.OptionalStringArg(args, "maxAmountPerPayment")),
		Windows:             windows,
		Method:              utils.OptionalStringArg(args, "method"),
		Config:              config,
	}, nil
}
