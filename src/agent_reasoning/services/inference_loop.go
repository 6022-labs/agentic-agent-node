package services

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/6022-labs/agentic-agent-node/src/agent_reasoning/models"
	"github.com/6022-labs/agentic-agent-node/src/agent_reasoning/services/interfaces"
	"github.com/6022-labs/agentic-agent-node/src/agent_reasoning/settings"
	"github.com/6022-labs/agentic-agent-node/src/agent_reasoning/utils"
	runtime_utils "github.com/6022-labs/agentic-agent-node/src/runtime_common/utils"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

// InferenceLoop drives the LLM ↔ MCP exchange until the LLM stops emitting
// tool calls or MaxIterations is hit.
type InferenceLoop struct {
	logger       *zap.Logger
	llmExecutor  interfaces.LLMExecutorInterface
	mcpRequester interfaces.McpRequesterInterface
	settings     *settings.InferenceLoopSettings
}

func NewInferenceLoop(
	logger *zap.Logger,
	llmExecutor interfaces.LLMExecutorInterface,
	mcpRequester interfaces.McpRequesterInterface,
	settings *settings.InferenceLoopSettings,
) *InferenceLoop {
	return &InferenceLoop{
		logger:       logger,
		llmExecutor:  llmExecutor,
		mcpRequester: mcpRequester,
		settings:     settings,
	}
}

// Run returns the full message stream including assistant + tool turns added
// during the loop.
func (l *InferenceLoop) Run(
	ctx context.Context,
	mcpConnections []*models.MCPConnection,
	messages []*interfaces.ChatCompletionMessage,
) ([]*interfaces.ChatCompletionMessage, error) {
	toolsByConnectionId, err := l.listToolsByConnection(mcpConnections)
	if err != nil {
		return nil, err
	}

	for iteration := 0; iteration < l.settings.MaxIterations; iteration++ {
		completion, err := l.llmExecutor.Complete(ctx, mcpConnections, toolsByConnectionId, messages)
		if err != nil {
			return nil, fmt.Errorf("llm executor failed at iteration %d: %w", iteration, err)
		}

		assistantMessage := &interfaces.ChatCompletionMessage{
			Role:      "assistant",
			ToolCalls: completion.ToolCalls,
		}
		if completion.Text != nil {
			assistantMessage.Content = runtime_utils.DemojizeContent(*completion.Text)
		}
		messages = append(messages, assistantMessage)

		if len(completion.ToolCalls) == 0 {
			return messages, nil
		}

		toolMessages, err := l.dispatchToolCalls(completion.ToolCalls, mcpConnections, toolsByConnectionId)
		if err != nil {
			return nil, err
		}
		messages = append(messages, toolMessages...)
	}

	return nil, fmt.Errorf("%w: %d", interfaces.ErrInferenceLoopMaxIterationsExceeded, l.settings.MaxIterations)
}

func (l *InferenceLoop) listToolsByConnection(
	mcpConnections []*models.MCPConnection,
) (map[uuid.UUID][]*interfaces.McpTool, error) {
	toolsByConnectionId := make(map[uuid.UUID][]*interfaces.McpTool, len(mcpConnections))
	for _, conn := range mcpConnections {
		tools, err := l.mcpRequester.ListTools(conn)
		if err != nil {
			return nil, fmt.Errorf("failed to list tools for mcp connection %q: %w", conn.Name, err)
		}
		toolsByConnectionId[conn.Id] = tools
	}
	return toolsByConnectionId, nil
}

func (l *InferenceLoop) dispatchToolCalls(
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

		result, err := l.mcpRequester.CallTool(mcpConnection, mcpTool, tc.Arguments)
		if err != nil {
			return nil, fmt.Errorf("failed to call mcp tool %q: %w", tc.Name, err)
		}

		resultBytes, err := json.Marshal(result)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal mcp tool result for %q: %w", tc.Name, err)
		}

		out = append(out, &interfaces.ChatCompletionMessage{
			Role:       "tool",
			ToolCallId: tc.Id,
			Content:    string(resultBytes),
		})
	}

	return out, nil
}
