package interfaces

import (
	"context"
	"errors"

	"github.com/6022-labs/agentic-agent-node/src/agent_reasoning/models"
)

// ErrInferenceLoopMaxIterationsExceeded — loop hit MaxIterations.
var ErrInferenceLoopMaxIterationsExceeded = errors.New("inner loop exceeded max iterations")

type InferenceLoopInterface interface {
	Run(
		ctx context.Context,
		mcpConnections []*models.MCPConnection,
		messages []*ChatCompletionMessage,
	) ([]*ChatCompletionMessage, error)
}
