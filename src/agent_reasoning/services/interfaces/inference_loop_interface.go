package interfaces

import (
	"context"
	"errors"

	"github.com/6022-labs/agentic-agent-node/src/agent_reasoning/models"
	"github.com/google/uuid"
)

// ErrInferenceLoopMaxIterationsExceeded — loop hit MaxIterations.
var ErrInferenceLoopMaxIterationsExceeded = errors.New("inner loop exceeded max iterations")

// ErrInferenceLoopNoActionableOutput — a turn produced no text, no tool calls, and no provider-run tool calls.
var ErrInferenceLoopNoActionableOutput = errors.New("inner loop received no actionable output")

type InferenceLoopInterface interface {
	// Run drives the LLM ↔ MCP exchange over pre-listed tools until the LLM
	// stops emitting tool calls or MaxIterations is hit.
	Run(
		ctx context.Context,
		mcpConnections []*models.MCPConnection,
		mcpToolsByConnectionId map[uuid.UUID][]*McpTool,
		messages []*ChatCompletionMessage,
	) ([]*ChatCompletionMessage, error)
}
