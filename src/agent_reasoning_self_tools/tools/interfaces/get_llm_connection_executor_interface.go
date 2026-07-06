package interfaces

import (
	"context"
	reasoning_responses "github.com/6022-labs/agentic-agent-node/src/agent_reasoning/use_cases/responses"
)

// GetLLMConnectionExecutorInterface exposes the LLM connection read (never the api key) to the self tools.
type GetLLMConnectionExecutorInterface interface {
	Execute(ctx context.Context) (*reasoning_responses.LLMConnectionResponse, error)
}
