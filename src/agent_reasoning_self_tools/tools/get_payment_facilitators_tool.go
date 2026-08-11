package tools

import (
	"context"

	"github.com/6022-labs/agentic-agent-node/src/agent_reasoning/models"
	tool_interfaces "github.com/6022-labs/agentic-agent-node/src/agent_reasoning_self_tools/tools/interfaces"
	"github.com/6022-labs/agentic-agent-node/src/agent_reasoning_self_tools/utils"
)

// GetPaymentFacilitatorsTool exposes the GetPaymentFacilitators use case as a built-in self tool.
type GetPaymentFacilitatorsTool struct {
	getPaymentFacilitatorsExecutor tool_interfaces.GetPaymentFacilitatorsExecutorInterface
}

func NewGetPaymentFacilitatorsTool(
	getPaymentFacilitatorsExecutor tool_interfaces.GetPaymentFacilitatorsExecutorInterface,
) *GetPaymentFacilitatorsTool {
	return &GetPaymentFacilitatorsTool{getPaymentFacilitatorsExecutor: getPaymentFacilitatorsExecutor}
}

func (t *GetPaymentFacilitatorsTool) Definition() *models.Tool {
	return &models.Tool{
		Name:        "get_payment_facilitators",
		Description: "Read the configured x402 payment facilitators per network.",
		InputScheme: utils.EmptyObjectScheme(),
	}
}

func (t *GetPaymentFacilitatorsTool) Execute(ctx context.Context, _ map[string]any) (map[string]any, error) {
	return utils.ToToolResult(t.getPaymentFacilitatorsExecutor.Execute(ctx))
}

func (t *GetPaymentFacilitatorsTool) IsEnabledByDefault() bool {
	return false
}
