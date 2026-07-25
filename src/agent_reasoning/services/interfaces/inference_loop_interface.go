package interfaces

import (
	"context"
	"errors"

	"github.com/6022-labs/agentic-agent-node/src/agent_reasoning/models"
)

// ErrInferenceLoopMaxIterationsExceeded — loop hit MaxIterations.
var ErrInferenceLoopMaxIterationsExceeded = errors.New("inner loop exceeded max iterations")

// ErrInferenceLoopDeadlineExceeded — loop hit the MaxDurationSeconds wall-clock ceiling.
var ErrInferenceLoopDeadlineExceeded = errors.New("inner loop exceeded max duration")

// ErrInferenceLoopTokenBudgetExceeded — cumulative LLM usage passed MaxTotalTokens.
var ErrInferenceLoopTokenBudgetExceeded = errors.New("inner loop exceeded token budget")

// ErrInferenceLoopNoActionableOutput — a turn produced no text, no tool calls, and no provider-run tool calls.
var ErrInferenceLoopNoActionableOutput = errors.New("inner loop received no actionable output")

type InferenceLoopInterface interface {
	// Run drives the LLM ↔ tool exchange (MCP and built-in self tools) until
	// the LLM stops emitting tool calls or a safety ceiling (iterations, wall-clock, tokens) is hit.
	Run(
		ctx context.Context,
		llmConnection *models.LLMConnection,
		selfTools []*models.Tool,
		mcpTools map[*models.MCPConnection][]*models.Tool,
		messages []*ChatCompletionMessage,
	) ([]*ChatCompletionMessage, error)
}
