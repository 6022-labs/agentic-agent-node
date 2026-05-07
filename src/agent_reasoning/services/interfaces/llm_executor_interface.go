package interfaces

import (
	"context"
	"errors"

	"github.com/6022-labs/agentic-agent-node/src/agent_reasoning/models"
	"github.com/google/uuid"
)

// ErrLLMConnectionNotConfigured is returned by LLMExecutor.Execute when no
// LLM connection has been set up for this agent yet.
var ErrLLMConnectionNotConfigured = errors.New("llm connection not configured")

// LLMExecutorInterface runs a single LLM round-trip. It does not loop on tool
// calls — the inner-loop service handles tool resolution and re-invocation.
//
// The caller passes the MCP connections + their already-listed tools so the
// executor can hand them to the underlying chat-completion adapter without
// re-listing per call.
type LLMExecutorInterface interface {
	Execute(
		ctx context.Context,
		mcpConnections []*models.MCPConnection,
		mcpToolsByConnectionId map[uuid.UUID][]*McpTool,
		messages []*ChatCompletionMessage,
	) (*LLMCompletion, error)
}
