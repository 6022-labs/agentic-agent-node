package interfaces

import (
	"context"

	"github.com/6022-labs/agentic-agent-node/src/agent_reasoning/models"
)

// LLMConnectionValidatorInterface checks a connection can actually complete
// before it is persisted; an empty apiKey means "use the stored one".
type LLMConnectionValidatorInterface interface {
	Validate(ctx context.Context, llmConnection *models.LLMConnection, apiKey string) error
}
