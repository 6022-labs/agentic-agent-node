package tools

import (
	"context"

	"github.com/6022-labs/agentic-agent-node/src/agent_reasoning/models"
	tool_interfaces "github.com/6022-labs/agentic-agent-node/src/agent_reasoning_self_tools/tools/interfaces"
	"github.com/6022-labs/agentic-agent-node/src/agent_reasoning_self_tools/utils"
)

// GetBlockchainConfigTool exposes the GetBlockchainConfig use case as a built-in self tool.
type GetBlockchainConfigTool struct {
	getBlockchainConfigExecutor tool_interfaces.GetBlockchainConfigExecutorInterface
}

func NewGetBlockchainConfigTool(getBlockchainConfigExecutor tool_interfaces.GetBlockchainConfigExecutorInterface) *GetBlockchainConfigTool {
	return &GetBlockchainConfigTool{getBlockchainConfigExecutor: getBlockchainConfigExecutor}
}

func (t *GetBlockchainConfigTool) Definition() *models.Tool {
	return &models.Tool{
		Name:        "get_blockchain_config",
		Description: "Read the operator blockchain configuration (chains, contracts, token).",
		InputScheme: utils.EmptyObjectScheme(),
	}
}

func (t *GetBlockchainConfigTool) Execute(ctx context.Context, _ map[string]any) (map[string]any, error) {
	return utils.ToToolResult(t.getBlockchainConfigExecutor.Execute(ctx))
}

func (t *GetBlockchainConfigTool) IsEnabledByDefault() bool {
	return false
}
