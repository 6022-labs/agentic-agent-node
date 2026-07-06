package interfaces

import (
	"context"
	reasoning_responses "github.com/6022-labs/agentic-agent-node/src/agent_reasoning/use_cases/responses"
)

// ListMCPConnectionsExecutorInterface exposes the MCP connections list (header names only) to the self tools.
type ListMCPConnectionsExecutorInterface interface {
	Execute(ctx context.Context) (*reasoning_responses.MCPConnectionListResponse, error)
}
