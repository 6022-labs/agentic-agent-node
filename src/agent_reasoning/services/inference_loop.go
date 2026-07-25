package services

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	metrics_interfaces "github.com/6022-labs/agentic-agent-node/src/agent_reasoning/metrics/interfaces"
	"github.com/6022-labs/agentic-agent-node/src/agent_reasoning/models"
	"github.com/6022-labs/agentic-agent-node/src/agent_reasoning/services/interfaces"
	"github.com/6022-labs/agentic-agent-node/src/agent_reasoning/settings"
	traces_interfaces "github.com/6022-labs/agentic-agent-node/src/agent_reasoning/traces/interfaces"
	"github.com/6022-labs/agentic-agent-node/src/agent_reasoning/utils"
	apperrors "github.com/6022-labs/agentic-agent-node/src/common/errors"
	runtime_services_interfaces "github.com/6022-labs/agentic-agent-node/src/runtime_common/services/interfaces"
	"go.uber.org/zap"
)

// InferenceLoop drives the LLM ↔ MCP exchange until the LLM stops emitting
// tool calls or a safety ceiling (iterations, wall-clock, tokens) is hit.
type InferenceLoop struct {
	logger                *zap.Logger
	llmExecutor           interfaces.LLMExecutorInterface
	mcpRequester          interfaces.McpRequesterInterface
	selfToolRegistry      interfaces.SelfToolRegistryInterface
	inferenceLoopSettings *settings.InferenceLoopSettings
	reasoningMetrics      metrics_interfaces.ReasoningMetricsInterface
	reasoningTracer       traces_interfaces.ReasoningTracerInterface
	emojiDecoder          runtime_services_interfaces.EmojiDecoderInterface
}

func NewInferenceLoop(
	logger *zap.Logger,
	llmExecutor interfaces.LLMExecutorInterface,
	mcpRequester interfaces.McpRequesterInterface,
	selfToolRegistry interfaces.SelfToolRegistryInterface,
	inferenceLoopSettings *settings.InferenceLoopSettings,
	reasoningMetrics metrics_interfaces.ReasoningMetricsInterface,
	reasoningTracer traces_interfaces.ReasoningTracerInterface,
	emojiDecoder runtime_services_interfaces.EmojiDecoderInterface,
) *InferenceLoop {
	return &InferenceLoop{
		logger:                logger,
		llmExecutor:           llmExecutor,
		mcpRequester:          mcpRequester,
		selfToolRegistry:      selfToolRegistry,
		inferenceLoopSettings: inferenceLoopSettings,
		reasoningMetrics:      reasoningMetrics,
		reasoningTracer:       reasoningTracer,
		emojiDecoder:          emojiDecoder,
	}
}

// Run returns the full message stream including assistant + tool turns added
// during the loop.
func (l *InferenceLoop) Run(
	ctx context.Context,
	llmConnection *models.LLMConnection,
	selfTools []*models.Tool,
	mcpTools map[*models.MCPConnection][]*models.Tool,
	messages []*interfaces.ChatCompletionMessage,
) ([]*interfaces.ChatCompletionMessage, error) {
	start := time.Now()

	ctx, cancel := context.WithTimeout(
		ctx,
		time.Duration(l.inferenceLoopSettings.MaxDurationSeconds)*time.Second,
	)
	defer cancel()

	totalTokens := 0

	for iteration := 0; iteration < l.inferenceLoopSettings.MaxIterations; iteration++ {
		// Checked before the next round-trip, not after, so a final answer that lands on the budget edge is never discarded.
		if totalTokens > l.inferenceLoopSettings.MaxTotalTokens {
			l.reasoningMetrics.RecordLoopAborted(ctx, "token_budget_exceeded", time.Since(start), int64(iteration))
			return nil, fmt.Errorf("%w: %d tokens used", interfaces.ErrInferenceLoopTokenBudgetExceeded, totalTokens)
		}

		completion, err := l.llmExecutor.Complete(ctx, llmConnection, selfTools, mcpTools, messages)
		if err != nil {
			if errors.Is(err, context.DeadlineExceeded) {
				l.reasoningMetrics.RecordLoopAborted(ctx, "deadline_exceeded", time.Since(start), int64(iteration))
				return nil, fmt.Errorf("%w after %s", interfaces.ErrInferenceLoopDeadlineExceeded, time.Since(start).Round(time.Millisecond))
			}
			return nil, fmt.Errorf("llm executor failed at iteration %d: %w", iteration, err)
		}

		if completion.Usage != nil {
			totalTokens += completion.Usage.InputTokens + completion.Usage.OutputTokens
		}

		messages = append(
			messages,
			l.recordExecutedToolCalls(ctx, completion.ExecutedToolCalls)...,
		)

		assistantText := l.selectAssistantText(completion.Texts)

		if assistantText != nil || len(completion.ToolCalls) > 0 {
			assistantMessage := &interfaces.ChatCompletionMessage{
				Role:      "assistant",
				ToolCalls: completion.ToolCalls,
			}
			if assistantText != nil {
				assistantMessage.Content = l.emojiDecoder.DemojizeContent(*assistantText)
			}
			messages = append(messages, assistantMessage)
		}

		if len(completion.ToolCalls) > 0 {
			l.reasoningMetrics.RecordToolCalls(ctx, int64(len(completion.ToolCalls)))
			toolMessages, err := l.dispatchToolCalls(ctx, completion.ToolCalls, mcpTools)
			if err != nil {
				return nil, err
			}
			messages = append(messages, toolMessages...)
			continue
		}

		if assistantText != nil {
			l.reasoningMetrics.RecordLoopCompleted(ctx, time.Since(start), int64(iteration+1))
			return messages, nil
		}

		// Provider ran tools server-side but hasn't answered — loop again for the text.
		if len(completion.ExecutedToolCalls) > 0 {
			continue
		}

		return nil, fmt.Errorf("%w at iteration %d", interfaces.ErrInferenceLoopNoActionableOutput, iteration)
	}

	l.reasoningMetrics.RecordLoopExceeded(ctx, time.Since(start), int64(l.inferenceLoopSettings.MaxIterations))

	return nil, fmt.Errorf("%w: %d", interfaces.ErrInferenceLoopMaxIterationsExceeded, l.inferenceLoopSettings.MaxIterations)
}

// selectAssistantText returns the last non-empty text block (earlier blocks are pre-tool-call preamble), or nil.
func (l *InferenceLoop) selectAssistantText(texts []string) *string {
	for i := len(texts) - 1; i >= 0; i-- {
		if texts[i] != "" {
			t := texts[i]
			return &t
		}
	}
	return nil
}

// recordExecutedToolCalls mirrors provider-run MCP calls into the stream (assistant turn + a tool turn per result) so they persist without being re-dispatched.
func (l *InferenceLoop) recordExecutedToolCalls(
	ctx context.Context,
	executed []interfaces.ExecutedToolCall,
) []*interfaces.ChatCompletionMessage {
	if len(executed) == 0 {
		return nil
	}

	toolCalls := make([]interfaces.ToolCall, 0, len(executed))
	for _, e := range executed {
		toolCalls = append(
			toolCalls,
			interfaces.ToolCall{
				Id:        e.Id,
				Name:      e.Name,
				Arguments: e.Arguments,
			},
		)
	}

	out := make([]*interfaces.ChatCompletionMessage, 0, len(executed)+1)
	out = append(
		out,
		&interfaces.ChatCompletionMessage{
			Role:      "assistant",
			ToolCalls: toolCalls,
		},
	)
	for _, e := range executed {
		out = append(
			out,
			&interfaces.ChatCompletionMessage{
				Role:       "tool",
				ToolCallId: e.Id,
				Content:    e.Output,
			},
		)
	}

	l.reasoningMetrics.RecordToolCalls(ctx, int64(len(executed)))

	return out
}

func (l *InferenceLoop) dispatchToolCalls(
	ctx context.Context,
	toolCalls []interfaces.ToolCall,
	mcpTools map[*models.MCPConnection][]*models.Tool,
) ([]*interfaces.ChatCompletionMessage, error) {
	out := make([]*interfaces.ChatCompletionMessage, 0, len(toolCalls))

	for _, toolCall := range toolCalls {
		message, err := l.dispatchToolCall(ctx, toolCall, mcpTools)
		if err != nil {
			l.logger.Warn("tool call failed", zap.String("toolName", toolCall.Name), zap.Error(err))
			message, err = l.toolResultMessage(toolCall, l.errorPayload(err))
			if err != nil {
				return nil, err
			}
		}
		out = append(out, message)
	}

	return out, nil
}

func (l *InferenceLoop) dispatchToolCall(
	ctx context.Context,
	toolCall interfaces.ToolCall,
	mcpTools map[*models.MCPConnection][]*models.Tool,
) (*interfaces.ChatCompletionMessage, error) {
	namespace := utils.ExtractNamespaceFromFunctionName(toolCall.Name)
	switch namespace {
	case utils.SelfToolNamespace:
		selfToolName, err := utils.ExtractSelfToolNameFromFunctionName(toolCall.Name)
		if err != nil {
			return nil, apperrors.NewValidationError("tool", fmt.Sprintf("failed to map tool call %q: %s", toolCall.Name, err.Error()))
		}

		l.logger.Debug("Calling self tool", zap.String("selfToolName", selfToolName))

		return l.callSelfTool(ctx, toolCall, selfToolName)

	case utils.McpToolNamespace:
		mcpConnection, mcpTool, err := utils.ExtractMcpInfoFromFunctionName(toolCall.Name, mcpTools)
		if err != nil {
			return nil, apperrors.NewValidationError("tool", fmt.Sprintf("failed to map tool call %q: %s", toolCall.Name, err.Error()))
		}

		l.logger.Debug("Calling MCP tool",
			zap.String("mcpConnectionName", mcpConnection.Name),
			zap.String("mcpToolName", mcpTool.Name),
		)

		return l.callMcpTool(ctx, toolCall, mcpConnection, mcpTool)

	default:
		return nil, apperrors.NewValidationError("tool", fmt.Sprintf("failed to map tool call %q: unknown tool namespace %q", toolCall.Name, namespace))
	}
}

func (l *InferenceLoop) callSelfTool(
	ctx context.Context,
	toolCall interfaces.ToolCall,
	selfToolName string,
) (*interfaces.ChatCompletionMessage, error) {
	ctx, span := l.reasoningTracer.StartToolCall(ctx, selfToolName, toolCall.Id, utils.SelfToolNamespace)
	defer span.End()

	result, err := l.selfToolRegistry.CallTool(ctx, selfToolName, toolCall.Arguments)
	if err != nil {
		span.Fail(err)
		return nil, err
	}
	return l.toolResultMessage(toolCall, result)
}

func (l *InferenceLoop) callMcpTool(
	ctx context.Context,
	toolCall interfaces.ToolCall,
	mcpConnection *models.MCPConnection,
	mcpTool *models.Tool,
) (*interfaces.ChatCompletionMessage, error) {
	ctx, span := l.reasoningTracer.StartToolCall(ctx, mcpTool.Name, toolCall.Id, mcpConnection.Name)
	defer span.End()

	result, err := l.mcpRequester.CallTool(ctx, mcpConnection, mcpTool, toolCall.Arguments)
	if err != nil {
		span.Fail(err)
		return nil, err
	}
	return l.toolResultMessage(toolCall, result)
}

func (l *InferenceLoop) toolResultMessage(toolCall interfaces.ToolCall, result map[string]any) (*interfaces.ChatCompletionMessage, error) {
	resultBytes, err := json.Marshal(result)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal tool result for %q: %w", toolCall.Name, err)
	}

	return &interfaces.ChatCompletionMessage{
		Role:       "tool",
		ToolCallId: toolCall.Id,
		Content:    string(resultBytes),
	}, nil
}

// errorPayload sanitizes a failed tool call for the model: business-error messages pass through, but internal/unavailable/unknown errors surface only a safe code with a generic message so infra detail (DB, Vault, HTTP) never reaches the model or transcript.
func (l *InferenceLoop) errorPayload(err error) map[string]any {
	var validationErr *apperrors.ValidationError
	if errors.As(err, &validationErr) {
		return l.errObject(map[string]any{"field": validationErr.Field, "message": validationErr.Message})
	}
	var notFoundErr *apperrors.NotFoundError
	if errors.As(err, &notFoundErr) {
		return l.errObject(map[string]any{"code": notFoundErr.Code, "message": notFoundErr.Message})
	}
	var conflictErr *apperrors.ConflictError
	if errors.As(err, &conflictErr) {
		return l.errObject(map[string]any{"code": conflictErr.Code, "message": conflictErr.Message})
	}
	var unauthorizedErr *apperrors.UnauthorizedError
	if errors.As(err, &unauthorizedErr) {
		return l.errObject(map[string]any{"code": unauthorizedErr.Code, "message": unauthorizedErr.Message})
	}
	var internalErr *apperrors.InternalError
	if errors.As(err, &internalErr) {
		return l.errObject(map[string]any{"code": internalErr.Code, "message": "the tool encountered an internal error"})
	}
	var unavailableErr *apperrors.UnavailableError
	if errors.As(err, &unavailableErr) {
		return l.errObject(map[string]any{"code": unavailableErr.Code, "message": "the tool is temporarily unavailable"})
	}
	return l.errObject(map[string]any{"message": "the tool call failed"})
}

func (*InferenceLoop) errObject(details map[string]any) map[string]any {
	return map[string]any{"error": details}
}
