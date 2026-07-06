package interfaces

import (
	"context"
	"errors"

	"github.com/6022-labs/agentic-agent-node/src/agent_reasoning/models"
)

// ErrInferenceLoopMaxIterationsExceeded — loop hit MaxIterations.
var ErrInferenceLoopMaxIterationsExceeded = errors.New("inner loop exceeded max iterations")

// ErrInferenceLoopNoActionableOutput — a turn produced no text, no tool calls, and no provider-run tool calls.
var ErrInferenceLoopNoActionableOutput = errors.New("inner loop received no actionable output")

type InferenceLoopInterface interface {
	// Run drives the LLM ↔ tool exchange (MCP and built-in self tools) until
	// the LLM stops emitting tool calls or MaxIterations is hit.
	Run(
		ctx context.Context,
		selfTools []*models.Tool,
		mcpTools map[*models.MCPConnection][]*models.Tool,
		messages []*ChatCompletionMessage,
	) ([]*ChatCompletionMessage, error)
}
