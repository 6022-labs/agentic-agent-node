package interfaces

import (
	"context"

	reasoning_requests "github.com/6022-labs/agentic-agent-node/src/agent_reasoning/use_cases/requests"
	reasoning_responses "github.com/6022-labs/agentic-agent-node/src/agent_reasoning/use_cases/responses"
)

// CallAgentExecutorInterface exposes one outbound A2A turn to the self tools.
type CallAgentExecutorInterface interface {
	Execute(
		ctx context.Context,
		request *reasoning_requests.CallAgentRequest,
	) (*reasoning_responses.CallAgentResponse, error)
}
