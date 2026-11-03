package tools

import (
	"context"

	payment_requests "github.com/6022-labs/agentic-agent-node/src/agent_payment/use_cases/requests"
	"github.com/6022-labs/agentic-agent-node/src/agent_reasoning/models"
	tool_interfaces "github.com/6022-labs/agentic-agent-node/src/agent_reasoning_self_tools/tools/interfaces"
	"github.com/6022-labs/agentic-agent-node/src/agent_reasoning_self_tools/utils"
	"github.com/ethereum/go-ethereum/signer/core/apitypes"
)

// SignTypedDataTool exposes raw EIP-712 signing; off by default since the payload is unconstrained.
type SignTypedDataTool struct {
	signTypedDataExecutor tool_interfaces.SignTypedDataExecutorInterface
}

func NewSignTypedDataTool(
	signTypedDataExecutor tool_interfaces.SignTypedDataExecutorInterface,
) *SignTypedDataTool {
	return &SignTypedDataTool{signTypedDataExecutor: signTypedDataExecutor}
}

func (t *SignTypedDataTool) Definition() *models.Tool {
	return &models.Tool{
		Name: "sign_typed_data",
		Description: "Sign EIP-712 typed data with your wallet key. Token transfers and permits count against " +
			"your spend policy; batch permits are refused. Prefer sign_x402_payment to pay an agent.",
		InputScheme: utils.ObjectScheme(map[string]models.ToolInputSchemeProperty{
			"typedData": utils.ObjectProperty("The EIP-712 payload: types, primaryType, domain and message."),
		}, []string{"typedData"}),
	}
}

func (t *SignTypedDataTool) Execute(ctx context.Context, args map[string]any) (map[string]any, error) {
	typedData, err := utils.JsonObjectArg[apitypes.TypedData](args, "typedData")
	if err != nil {
		return nil, err
	}
	return utils.ToToolResult(
		t.signTypedDataExecutor.Execute(ctx, &payment_requests.SignTypedDataRequest{TypedData: typedData}),
	)
}

func (t *SignTypedDataTool) IsEnabledByDefault() bool {
	return false
}

func (t *SignTypedDataTool) GroupName() string {
	return GroupPayments
}
