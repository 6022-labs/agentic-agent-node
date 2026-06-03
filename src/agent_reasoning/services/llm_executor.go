package services

import (
	"context"
	"fmt"

	"github.com/6022-labs/agentic-agent-node/src/agent_reasoning/models"
	"github.com/6022-labs/agentic-agent-node/src/agent_reasoning/repositories"
	"github.com/6022-labs/agentic-agent-node/src/agent_reasoning/services/interfaces"
	"github.com/google/uuid"
)

// LLMExecutor runs one chat-completion round-trip. No looping — InferenceLoop
// owns that.
type LLMExecutor struct {
	llmConnectionRepository repositories.LLMConnectionRepositoryInterface
	chatCompletionRequester interfaces.ChatCompletionRequesterInterface
}

func NewLLMExecutor(
	llmConnectionRepository repositories.LLMConnectionRepositoryInterface,
	chatCompletionRequester interfaces.ChatCompletionRequesterInterface,
) *LLMExecutor {
	return &LLMExecutor{
		llmConnectionRepository: llmConnectionRepository,
		chatCompletionRequester: chatCompletionRequester,
	}
}

func (e *LLMExecutor) Complete(
	ctx context.Context,
	mcpConnections []*models.MCPConnection,
	mcpToolsByConnectionId map[uuid.UUID][]*interfaces.McpTool,
	messages []*interfaces.ChatCompletionMessage,
) (*interfaces.LLMCompletion, error) {
	llmConnection, err := e.llmConnectionRepository.Get(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to load llm connection: %w", err)
	}
	if llmConnection == nil {
		return nil, interfaces.ErrLLMConnectionNotConfigured
	}

	return e.chatCompletionRequester.ChatCompletion(
		ctx,
		llmConnection,
		mcpConnections,
		mcpToolsByConnectionId,
		messages,
	)
}
