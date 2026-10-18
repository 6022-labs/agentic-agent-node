package interfaces

import (
	"context"

	"github.com/6022-labs/agentic-agent-node/src/agent_reasoning/models"
)

// SelfToolRegistryInterface aggregates built-in tools under the reserved "self" connection.
type SelfToolRegistryInterface interface {
	// ListTools returns the definitions of every registered self tool.
	ListTools() []*models.Tool

	// CallTool runs a registered tool; ctx carries the turn's caller id (helpers.CallerIdFromContext).
	CallTool(ctx context.Context, toolName string, args map[string]any) (map[string]any, error)

	// ListDefaultEnabled maps every registered tool name to its declared default enabled flag.
	ListDefaultEnabled() map[string]bool

	// ListGroupNames maps every registered tool name to its console display group.
	ListGroupNames() map[string]string
}
