package interfaces

import (
	"context"
	reasoning_requests "github.com/6022-labs/agentic-agent-node/src/agent_reasoning/use_cases/requests"
	reasoning_responses "github.com/6022-labs/agentic-agent-node/src/agent_reasoning/use_cases/responses"
)

// SetAgentGuidanceExecutorInterface exposes the guidance write to the self tools.
type SetAgentGuidanceExecutorInterface interface {
	Execute(ctx context.Context, request *reasoning_requests.SetAgentGuidanceRequest) (*reasoning_responses.AgentGuidanceResponse, error)
}
