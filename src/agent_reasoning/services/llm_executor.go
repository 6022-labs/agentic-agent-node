package services

import (
	"context"
	"fmt"

	identity_repositories "github.com/6022-labs/agentic-agent-node/src/agent_identity/repositories"
	"github.com/6022-labs/agentic-agent-node/src/agent_reasoning/models"
	"github.com/6022-labs/agentic-agent-node/src/agent_reasoning/repositories"
	"github.com/6022-labs/agentic-agent-node/src/agent_reasoning/services/interfaces"
	traces_interfaces "github.com/6022-labs/agentic-agent-node/src/agent_reasoning/traces/interfaces"
)

// LLMExecutor runs one chat-completion round-trip. No looping — InferenceLoop
// owns that.
type LLMExecutor struct {
	llmConnectionRepository repositories.LLMConnectionRepositoryInterface
	chatCompletionRequester interfaces.ChatCompletionRequesterInterface
	agentRepository         identity_repositories.AgentRepositoryInterface
	reasoningTracer         traces_interfaces.ReasoningTracerInterface
}

func NewLLMExecutor(
	llmConnectionRepository repositories.LLMConnectionRepositoryInterface,
	chatCompletionRequester interfaces.ChatCompletionRequesterInterface,
	agentRepository identity_repositories.AgentRepositoryInterface,
	reasoningTracer traces_interfaces.ReasoningTracerInterface,
) *LLMExecutor {
	return &LLMExecutor{
		llmConnectionRepository: llmConnectionRepository,
		chatCompletionRequester: chatCompletionRequester,
		agentRepository:         agentRepository,
		reasoningTracer:         reasoningTracer,
	}
}

func (e *LLMExecutor) Complete(
	ctx context.Context,
	selfTools []*models.Tool,
	mcpTools map[*models.MCPConnection][]*models.Tool,
	messages []*interfaces.ChatCompletionMessage,
) (*interfaces.LLMCompletion, error) {
	agent, err := e.agentRepository.Get(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to load agent identity: %w", err)
	}
	if agent == nil {
		return nil, interfaces.ErrLLMConnectionNotConfigured
	}
	llmConnection, err := e.llmConnectionRepository.GetByAgentId(ctx, agent.Id)
	if err != nil {
		return nil, fmt.Errorf("failed to load llm connection: %w", err)
	}
	if llmConnection == nil {
		return nil, interfaces.ErrLLMConnectionNotConfigured
	}

	ctx, span := e.reasoningTracer.StartChatCompletion(ctx, string(llmConnection.Platform), llmConnection.Model)
	defer span.End()

	completion, err := e.chatCompletionRequester.ChatCompletion(
		ctx,
		llmConnection,
		selfTools,
		mcpTools,
		messages,
	)
	if err != nil {
		span.Fail(err)
		return nil, err
	}
	if completion.Usage != nil {
		span.RecordUsage(completion.Usage.InputTokens, completion.Usage.OutputTokens)
	}
	span.RecordResponse(completion.ResponseModel, completion.FinishReasons, completion.ResponseId)
	for _, executed := range completion.ExecutedToolCalls {
		span.RecordToolCall(executed.Name, executed.Id)
	}
	return completion, nil
}
