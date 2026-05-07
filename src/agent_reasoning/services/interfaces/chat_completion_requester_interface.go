package interfaces

import (
	"context"

	"github.com/6022-labs/agentic-agent-node/src/agent_reasoning/models"
	"github.com/google/uuid"
)

// ChatCompletionRequesterInterface dispatches a single chat-completion
// round-trip to the HTTP client matching the configured LLM platform. The api
// key is fetched internally via LLMSecretStorage.
//
// The adapter does NOT loop on tool calls. When the LLM emits tool requests,
// they are returned to the caller via LLMCompletion.ToolCalls; the
// inference-loop service is responsible for resolving them and re-calling.
type ChatCompletionRequesterInterface interface {
	ChatCompletion(
		ctx context.Context,
		llmConnection *models.LLMConnection,
		mcpConnections []*models.MCPConnection,
		mcpToolsByConnectionId map[uuid.UUID][]*McpTool,
		messages []*ChatCompletionMessage,
	) (*LLMCompletion, error)
}

// ChatCompletionMessage is one turn in the conversation passed to the LLM.
//
//   - Role "system" / "user": Content carries the text body.
//   - Role "assistant": Content holds the assistant text (may be empty when the
//     LLM only emitted tool calls); ToolCalls holds any tool-use requests from
//     the previous LLM turn so the next request can replay them.
//   - Role "tool": Content is the JSON-encoded tool result; ToolCallId points
//     back at the assistant ToolCall it answers.
type ChatCompletionMessage struct {
	Role       string     `json:"role"`
	Content    string     `json:"content"`
	ToolCalls  []ToolCall `json:"tool_calls,omitempty"`
	ToolCallId string     `json:"tool_call_id,omitempty"`
}

// LLMCompletion is the platform-agnostic result of a single chat-completion
// round-trip. Either Text is set (final answer or partial) or ToolCalls is
// non-empty (the LLM is asking the inference loop to run tools); both can be
// set when the platform emits text alongside tool calls.
type LLMCompletion struct {
	Text      *string
	ToolCalls []ToolCall
}

// ToolCall is a platform-agnostic tool-use request emitted by the LLM. The
// adapter maps it from the platform-specific shape (anthropic content
// type=tool_use, openai tool_calls, bedrock toolUse, etc.) so the inference
// loop can dispatch it without knowing which provider produced it.
//
// Name encodes the MCP connection + tool with the
// `<connection_name>_<tool_name>` convention from agent_reasoning/utils.
type ToolCall struct {
	Id        string
	Name      string
	Arguments map[string]any
}
