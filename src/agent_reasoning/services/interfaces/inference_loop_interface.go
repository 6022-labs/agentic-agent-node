package interfaces

import (
	"context"
	"errors"

	"github.com/6022-labs/agentic-agent-node/src/agent_reasoning/models"
)

// ErrInferenceLoopMaxIterationsExceeded is returned by Run when the loop hits the
// configured MaxIterations bound without a final assistant text.
var ErrInferenceLoopMaxIterationsExceeded = errors.New("inner loop exceeded max iterations")

// InferenceLoopInterface drives the LLM ↔ MCP exchange to completion. Each
// iteration sends the current message stream through the LLMExecutor; if the
// LLM emits ToolCalls, they are dispatched via McpRequester, results are
// appended as tool messages, and the loop iterates. The loop returns once the
// LLM produces a completion with no ToolCalls or when MaxIterations is hit.
//
// Use cases (e.g. create_response) depend on this interface, never on the
// concrete InferenceLoop service.
type InferenceLoopInterface interface {
	Run(
		ctx context.Context,
		mcpConnections []*models.MCPConnection,
		messages []*ChatCompletionMessage,
	) ([]*ChatCompletionMessage, error)
}
