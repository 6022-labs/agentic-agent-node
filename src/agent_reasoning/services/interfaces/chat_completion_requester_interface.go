package interfaces

import (
	"context"

	"github.com/6022-labs/agentic-agent-node/src/agent_reasoning/models"
)

// ChatCompletionRequesterInterface — single LLM round-trip. No tool-loop;
// tool resolution is the InferenceLoop's job.
type ChatCompletionRequesterInterface interface {
	ChatCompletion(
		ctx context.Context,
		llmConnection *models.LLMConnection,
		selfTools []*models.Tool,
		mcpTools map[*models.MCPConnection][]*models.Tool,
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

// LLMCompletion — platform-agnostic result of one round-trip. Texts holds every
// text block the provider emitted, in order (the loop selects the one to
// surface); Response* fields describe what the provider returned.
type LLMCompletion struct {
	Texts             []string
	ToolCalls         []ToolCall
	ExecutedToolCalls []ExecutedToolCall
	Usage             *LLMUsage
	ResponseModel     string
	FinishReasons     []string
	ResponseId        string
}

type LLMUsage struct {
	InputTokens  int
	OutputTokens int
}

// ToolCall — platform-agnostic tool-use request. Name uses the
// `<connection_name>_<tool_name>` convention from agent_reasoning/utils.
type ToolCall struct {
	Id        string
	Name      string
	Arguments map[string]any
}

// ExecutedToolCalls carry their own results — the provider ran them server-side, so the loop must not re-dispatch them.
type ExecutedToolCall struct {
	Id        string
	Name      string
	Arguments map[string]any
	Output    string
	IsError   bool
}
