package tools

import (
	"fmt"

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
		"{windowSeconds (integer, 86400 = a day), " +
			"maxAmount (non-negative integer string in atomic units; USDC has 6 decimals, so 1 USDC = 1000000)}.",
	)
	return map[string]models.ToolInputSchemeProperty{
		"name": utils.StringProperty("Rule name (1-64 characters)."),
		"maxAmountPerPayment": utils.StringProperty(
			"Optional cap on one payment in the asset's atomic units, e.g. 1000000 = 1 USDC (0 denies anything priced).",
		),
		"windows": utils.ArrayProperty(
			"Recurring budgets, at most one per windowSeconds; the rule needs a cap, a window, or both. Max 10.",
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
	maxAmountPerPayment, err := t.maxAmountPerPaymentArg(args)
	if err != nil {
		return nil, err
	}
	return &payment_requests.SavePaymentSpendRuleRequest{
		SavePaymentRuleRequest: payment_requests.SavePaymentRuleRequest{
			Name:   name,
			Method: utils.OptionalStringArg(args, "method"),
		},
		MaxAmountPerPayment: maxAmountPerPayment,
		Windows:             windows,
		Config:              config,
	}, nil
}

// maxAmountPerPaymentArg refuses a non-string cap, since reading it as absent would drop the cap.
func (t *AbstractPaymentSpendRuleTool) maxAmountPerPaymentArg(args map[string]any) (*string, error) {
	raw, found := args["maxAmountPerPayment"]
	if !found || raw == nil {
		return nil, nil
	}
	text, ok := raw.(string)
	if !ok {
		return nil, fmt.Errorf("argument %q must be a string", "maxAmountPerPayment")
	}
	return lo.EmptyableToPtr(text), nil
}
