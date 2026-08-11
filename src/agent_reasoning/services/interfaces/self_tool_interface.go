package interfaces

import (
	"context"

	"github.com/6022-labs/agentic-agent-node/src/agent_reasoning/models"
)

// SelfToolInterface is one built-in tool the agent can call on itself from the inference loop.
type SelfToolInterface interface {
	// Definition returns the tool's name, description and input schema as exposed to the LLM.
	Definition() *models.Tool

	// Execute runs the tool; domain failures come back as a result payload, not an error.
	Execute(ctx context.Context, args map[string]any) (map[string]any, error)

	// IsEnabledByDefault seeds the tool's enabled flag when its DB row is first created.
	IsEnabledByDefault() bool
}
