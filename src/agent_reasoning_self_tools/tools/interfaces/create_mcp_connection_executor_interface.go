package interfaces

import (
	"context"

	reasoning_requests "github.com/6022-labs/agentic-agent-node/src/agent_reasoning/use_cases/requests"
	reasoning_responses "github.com/6022-labs/agentic-agent-node/src/agent_reasoning/use_cases/responses"
)

// CreateMCPConnectionExecutorInterface exposes the MCP connection create to the self tools.
type CreateMCPConnectionExecutorInterface interface {
	Execute(ctx context.Context, request *reasoning_requests.SaveMCPConnectionRequest) (*reasoning_responses.MCPConnectionResponse, error)
}
