package services

import (
	"context"

	"github.com/6022-labs/agentic-agent-node/src/agent_reasoning/models"
	"github.com/6022-labs/agentic-agent-node/src/agent_reasoning/services/interfaces"
	traces_interfaces "github.com/6022-labs/agentic-agent-node/src/agent_reasoning/traces/interfaces"
)

// LLMExecutor runs one chat-completion round-trip. No looping — InferenceLoop
// owns that. The connection is resolved once per request by the caller.
type LLMExecutor struct {
	chatCompletionRequester interfaces.ChatCompletionRequesterInterface
	reasoningTracer         traces_interfaces.ReasoningTracerInterface
}

func NewLLMExecutor(
	chatCompletionRequester interfaces.ChatCompletionRequesterInterface,
	reasoningTracer traces_interfaces.ReasoningTracerInterface,
) *LLMExecutor {
	return &LLMExecutor{
		chatCompletionRequester: chatCompletionRequester,
		reasoningTracer:         reasoningTracer,
	}
}

func (e *LLMExecutor) Complete(
	ctx context.Context,
	llmConnection *models.LLMConnection,
	selfTools []*models.Tool,
	mcpTools map[*models.MCPConnection][]*models.Tool,
	messages []*interfaces.ChatCompletionMessage,
) (*interfaces.LLMCompletion, error) {
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
