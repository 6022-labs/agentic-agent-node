package tools

import (
	"context"
	"encoding/json"

	payment_requests "github.com/6022-labs/agentic-agent-node/src/agent_payment/use_cases/requests"
	"github.com/6022-labs/agentic-agent-node/src/agent_reasoning/models"
	tool_interfaces "github.com/6022-labs/agentic-agent-node/src/agent_reasoning_self_tools/tools/interfaces"
	"github.com/6022-labs/agentic-agent-node/src/agent_reasoning_self_tools/utils"
	"github.com/6022-labs/agentic-agent-node/src/common/errors"
)

// SignX402PaymentTool exposes the SignX402Payment use case as a built-in self tool.
type SignX402PaymentTool struct {
	signX402PaymentExecutor tool_interfaces.SignX402PaymentExecutorInterface
}

func NewSignX402PaymentTool(
	signX402PaymentExecutor tool_interfaces.SignX402PaymentExecutorInterface,
) *SignX402PaymentTool {
	return &SignX402PaymentTool{signX402PaymentExecutor: signX402PaymentExecutor}
}

func (t *SignX402PaymentTool) Definition() *models.Tool {
	return &models.Tool{
		Name: "sign_x402_payment",
		Description: "Pay an x402 quote from your wallet; returns the paymentSignature for the paid call. " +
			"Pass one accepts entry as received. Refused, with no spend, unless the payee is allowed and " +
			"the token's rule fits. Spend counts when signed, so don't sign again for the same call.",
		InputScheme: utils.ObjectScheme(map[string]models.ToolInputSchemeProperty{
			"accept": utils.ObjectProperty(
				"One accepts entry from a 402 quote: network, asset, payTo, amount and its extra.",
			),
		}, []string{"accept"}),
	}
}

func (t *SignX402PaymentTool) Execute(ctx context.Context, args map[string]any) (map[string]any, error) {
	raw, err := utils.RawJsonObjectArg(args, "accept")
	if err != nil {
		return nil, err
	}

	request := &payment_requests.SignX402PaymentRequest{}
	if err := json.Unmarshal(raw, request); err != nil {
		return nil, errors.NewValidationError("accept", "must be one x402 accepts entry, amount as a string")
	}
	return utils.ToToolResult(t.signX402PaymentExecutor.Execute(ctx, request))
}

func (t *SignX402PaymentTool) IsEnabledByDefault() bool {
	return false
}

func (t *SignX402PaymentTool) GroupName() string {
	return GroupPayments
}
