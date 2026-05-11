package interfaces

import (
	"context"
	"errors"

	"github.com/6022-labs/agentic-agent-node/src/agent_reasoning/models"
	"github.com/google/uuid"
)

// ErrLLMConnectionNotConfigured — no LLM connection set up for this agent.
var ErrLLMConnectionNotConfigured = errors.New("llm connection not configured")

type LLMExecutorInterface interface {
	Complete(
		ctx context.Context,
		mcpConnections []*models.MCPConnection,
		mcpToolsByConnectionId map[uuid.UUID][]*McpTool,
		messages []*ChatCompletionMessage,
	) (*LLMCompletion, error)
}
