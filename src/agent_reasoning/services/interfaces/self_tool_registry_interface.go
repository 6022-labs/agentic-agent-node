package interfaces

import (
	"context"

	"github.com/6022-labs/agentic-agent-node/src/agent_reasoning/models"
)

// SelfToolRegistryInterface aggregates built-in tools under the reserved "self" connection.
type SelfToolRegistryInterface interface {
	// ListTools returns the definitions of every registered self tool.
	ListTools() []*models.Tool

	// CallTool executes a registered tool and returns its result, or an error the caller surfaces to the model.
	CallTool(ctx context.Context, toolName string, args map[string]any) (map[string]any, error)
}
