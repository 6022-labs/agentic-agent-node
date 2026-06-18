package interfaces

import (
	"context"

	"github.com/6022-labs/agentic-agent-node/src/agent_reasoning/models"
	"github.com/google/uuid"
)

// ChatCompletionRequesterInterface — single LLM round-trip. No tool-loop;
// tool resolution is the InferenceLoop's job.
type ChatCompletionRequesterInterface interface {
	ChatCompletion(
		ctx context.Context,
		llmConnection *models.LLMConnection,
		mcpConnections []*models.MCPConnection,
		mcpToolsByConnectionId map[uuid.UUID][]*McpTool,
		messages []*ChatCompletionMessage,
	) (*LLMCompletion, error)
}

// ChatCompletionMessage — one turn passed to the LLM. Roles: system, user,
// assistant (Content + optional ToolCalls), tool (Content = JSON tool result,
// ToolCallId references the assistant ToolCall).
type ChatCompletionMessage struct {
	Role       string     `json:"role"`
	Content    string     `json:"content"`
	ToolCalls  []ToolCall `json:"tool_calls,omitempty"`
	ToolCallId string     `json:"tool_call_id,omitempty"`
}

// LLMCompletion — platform-agnostic result of one round-trip. Text and
// ToolCalls can both be non-empty when the LLM emits both.
type LLMCompletion struct {
	Text      *string
	ToolCalls []ToolCall
}

// ToolCall — platform-agnostic tool-use request. Name uses the
// `<connection_name>_<tool_name>` convention from agent_reasoning/utils.
type ToolCall struct {
	Id        string
	Name      string
	Arguments map[string]any
}
