package interfaces

import (
	"context"

	reasoning_requests "github.com/6022-labs/agentic-agent-node/src/agent_reasoning/use_cases/requests"
)

// DeleteMCPConnectionExecutorInterface exposes the MCP connection delete to the self tools.
type DeleteMCPConnectionExecutorInterface interface {
	Execute(ctx context.Context, request *reasoning_requests.DeleteMCPConnectionRequest) error
}
