package interfaces

import (
	"context"

	common_responses "github.com/6022-labs/agentic-agent-node/src/runtime_common/use_cases/responses"
)

// GetBlockchainConfigExecutorInterface exposes the blockchain config read to the self tools.
type GetBlockchainConfigExecutorInterface interface {
	Execute(ctx context.Context) (*common_responses.BlockchainConfigResponse, error)
}
