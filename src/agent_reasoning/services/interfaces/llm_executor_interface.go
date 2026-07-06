package interfaces

import (
	"context"
	"errors"

	"github.com/6022-labs/agentic-agent-node/src/agent_reasoning/models"
)

// ErrLLMConnectionNotConfigured — no LLM connection set up for this agent.
var ErrLLMConnectionNotConfigured = errors.New("llm connection not configured")

type LLMExecutorInterface interface {
	Complete(
		ctx context.Context,
		selfTools []*models.Tool,
		mcpTools map[*models.MCPConnection][]*models.Tool,
		messages []*ChatCompletionMessage,
	) (*LLMCompletion, error)
}
