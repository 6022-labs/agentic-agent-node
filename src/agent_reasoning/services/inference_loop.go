package services

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	metrics_interfaces "github.com/6022-labs/agentic-agent-node/src/agent_reasoning/metrics/interfaces"
	"github.com/6022-labs/agentic-agent-node/src/agent_reasoning/models"
	"github.com/6022-labs/agentic-agent-node/src/agent_reasoning/services/interfaces"
	"github.com/6022-labs/agentic-agent-node/src/agent_reasoning/settings"
	traces_interfaces "github.com/6022-labs/agentic-agent-node/src/agent_reasoning/traces/interfaces"
	"github.com/6022-labs/agentic-agent-node/src/agent_reasoning/utils"
	runtime_utils "github.com/6022-labs/agentic-agent-node/src/runtime_common/utils"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

// InferenceLoop drives the LLM ↔ MCP exchange until the LLM stops emitting
// tool calls or MaxIterations is hit.
type InferenceLoop struct {
	logger          *zap.Logger
	llmExecutor     interfaces.LLMExecutorInterface
	mcpRequester    interfaces.McpRequesterInterface
	settings        *settings.InferenceLoopSettings
	metrics         metrics_interfaces.ReasoningMetricsInterface
	reasoningTracer traces_interfaces.ReasoningTracerInterface
}

func NewInferenceLoop(
	logger *zap.Logger,
	llmExecutor interfaces.LLMExecutorInterface,
	mcpRequester interfaces.McpRequesterInterface,
	settings *settings.InferenceLoopSettings,
	metrics metrics_interfaces.ReasoningMetricsInterface,
	reasoningTracer traces_interfaces.ReasoningTracerInterface,
) *InferenceLoop {
	return &InferenceLoop{
		logger:          logger,
		llmExecutor:     llmExecutor,
		mcpRequester:    mcpRequester,
		settings:        settings,
		metrics:         metrics,
		reasoningTracer: reasoningTracer,
	}
}

// Run returns the full message stream including assistant + tool turns added
// during the loop.
func (l *InferenceLoop) Run(
	ctx context.Context,
	mcpConnections []*models.MCPConnection,
	toolsByConnectionId map[uuid.UUID][]*interfaces.McpTool,
	messages []*interfaces.ChatCompletionMessage,
) ([]*interfaces.ChatCompletionMessage, error) {
	start := time.Now()

	for iteration := 0; iteration < l.settings.MaxIterations; iteration++ {
		completion, err := l.llmExecutor.Complete(ctx, mcpConnections, toolsByConnectionId, messages)
		if err != nil {
			return nil, fmt.Errorf("llm executor failed at iteration %d: %w", iteration, err)
		}

		messages = append(
			messages,
			l.recordExecutedToolCalls(ctx, completion.ExecutedToolCalls)...,
		)

		if completion.Text != nil || len(completion.ToolCalls) > 0 {
			assistantMessage := &interfaces.ChatCompletionMessage{
				Role:      "assistant",
				ToolCalls: completion.ToolCalls,
			}
			if completion.Text != nil {
				assistantMessage.Content = runtime_utils.DemojizeContent(*completion.Text)
			}
			messages = append(messages, assistantMessage)
		}

		if len(completion.ToolCalls) > 0 {
			l.metrics.RecordToolCalls(ctx, int64(len(completion.ToolCalls)))
			toolMessages, err := l.dispatchToolCalls(ctx, completion.ToolCalls, mcpConnections, toolsByConnectionId)
			if err != nil {
				return nil, err
			}
			messages = append(messages, toolMessages...)
			continue
		}

		if completion.Text != nil {
			l.metrics.RecordLoopCompleted(ctx, time.Since(start), int64(iteration+1))
			return messages, nil
		}

		// Provider ran tools server-side but hasn't answered — loop again for the text.
		if len(completion.ExecutedToolCalls) > 0 {
			continue
		}

		return nil, fmt.Errorf("%w at iteration %d", interfaces.ErrInferenceLoopNoActionableOutput, iteration)
	}

	l.metrics.RecordLoopExceeded(ctx, time.Since(start), int64(l.settings.MaxIterations))

	return nil, fmt.Errorf("%w: %d", interfaces.ErrInferenceLoopMaxIterationsExceeded, l.settings.MaxIterations)
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

	l.metrics.RecordToolCalls(ctx, int64(len(executed)))

	return out
}

func (l *InferenceLoop) dispatchToolCalls(
	ctx context.Context,
	toolCalls []interfaces.ToolCall,
	mcpConnections []*models.MCPConnection,
	toolsByConnectionId map[uuid.UUID][]*interfaces.McpTool,
) ([]*interfaces.ChatCompletionMessage, error) {
	out := make([]*interfaces.ChatCompletionMessage, 0, len(toolCalls))

	for _, tc := range toolCalls {
		mcpConnection, mcpTool, err := utils.ExtractMcpInfoFromFunctionName(tc.Name, mcpConnections, toolsByConnectionId)
		if err != nil {
			return nil, fmt.Errorf("failed to map tool call %q: %w", tc.Name, err)
		}

		l.logger.Debug("Calling MCP tool",
			zap.String("mcpConnectionName", mcpConnection.Name),
			zap.String("mcpToolName", mcpTool.Name),
		)

		message, err := l.callTool(ctx, tc, mcpConnection, mcpTool)
		if err != nil {
			return nil, err
		}
		out = append(out, message)
	}

	return out, nil
}

func (l *InferenceLoop) callTool(
	ctx context.Context,
	tc interfaces.ToolCall,
	mcpConnection *models.MCPConnection,
	mcpTool *interfaces.McpTool,
) (*interfaces.ChatCompletionMessage, error) {
	_, span := l.reasoningTracer.StartToolCall(ctx, mcpTool.Name, tc.Id, mcpConnection.Name)
	defer span.End()

	result, err := l.mcpRequester.CallTool(mcpConnection, mcpTool, tc.Arguments)
	if err != nil {
		span.Fail(err)
		return nil, fmt.Errorf("failed to call mcp tool %q: %w", tc.Name, err)
	}

	resultBytes, err := json.Marshal(result)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal mcp tool result for %q: %w", tc.Name, err)
	}

	return &interfaces.ChatCompletionMessage{
		Role:       "tool",
		ToolCallId: tc.Id,
		Content:    string(resultBytes),
	}, nil
}
