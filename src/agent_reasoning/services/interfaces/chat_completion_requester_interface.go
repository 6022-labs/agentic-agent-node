package interfaces

import (
	"context"

	"github.com/6022-labs/agentic-agent-node/src/agent_reasoning/models"
)

// ChatCompletionRequesterInterface dispatches a chat-completion request to the
// HTTP client matching the configured LLM platform. The api key is fetched
// internally via LLMSecretStorage; callers don't need to know about Vault.
type ChatCompletionRequesterInterface interface {
	ChatCompletion(
		ctx context.Context,
		llmConnection *models.LLMConnection,
		mcpConnections []*models.MCPConnection,
		messages []*models.ChatCompletionMessage,
	) (*string, error)
}
