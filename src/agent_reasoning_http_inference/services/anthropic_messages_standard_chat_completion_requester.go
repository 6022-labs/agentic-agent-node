package services

import (
	"fmt"
	"strings"

	"github.com/6022-labs/agentic-agent-node/src/agent_reasoning/models"
	"github.com/6022-labs/agentic-agent-node/src/agent_reasoning/services/interfaces"
	"github.com/6022-labs/agentic-agent-node/src/agent_reasoning_http_inference/clients"
	external_models "github.com/6022-labs/agentic-agent-node/src/agent_reasoning_http_inference/models"
	"github.com/google/uuid"
	"github.com/samber/lo"
	"go.uber.org/zap"
)

type AnthropicMessagesStandardChatCompletionRequesterInterface interface {
	ChatCompletion(
		baseUrl, apiKey string,
		hasDirectAccessToMcp bool,
		llmConnection *models.LLMConnection,
		mcpConnections []*models.MCPConnection,
		mcpToolsByConnectionId map[uuid.UUID][]*interfaces.McpTool,
		messages []*interfaces.ChatCompletionMessage,
	) (*interfaces.LLMCompletion, error)
}

type AnthropicMessagesStandardChatCompletionRequester struct {
	logger                  *zap.Logger
	anthropicMessagesClient clients.AnthropicMessagesStandardClientInterface
}

func NewAnthropicMessagesStandardChatCompletionRequester(
	logger *zap.Logger,
	anthropicMessagesClient clients.AnthropicMessagesStandardClientInterface,
) *AnthropicMessagesStandardChatCompletionRequester {
	return &AnthropicMessagesStandardChatCompletionRequester{
		logger:                  logger,
		anthropicMessagesClient: anthropicMessagesClient,
	}
}

func (r *AnthropicMessagesStandardChatCompletionRequester) ChatCompletion(
	baseUrl, apiKey string,
	hasDirectAccessToMcp bool,
	llmConnection *models.LLMConnection,
	mcpConnections []*models.MCPConnection,
	mcpToolsByConnectionId map[uuid.UUID][]*interfaces.McpTool,
	messages []*interfaces.ChatCompletionMessage,
) (*interfaces.LLMCompletion, error) {
	systems := r.buildSystems(messages)
	extMessages, err := r.buildRequestMessages(messages)
	if err != nil {
		return nil, fmt.Errorf("failed to build anthropic messages: %w", err)
	}

	externalRequest := external_models.NewExternalAnthropicMessagesStandardRequest(
		llmConnection.Model,
		extMessages,
		systems,
	)

	directMcps := make(map[uuid.UUID]bool)
	if hasDirectAccessToMcp {
		// Anthropic only supports MCP servers with no auth or with a single OAuth Bearer Authorization header.
		for _, mcpConnection := range mcpConnections {
			auth, ok := mcpConnection.HeaderAuth["Authorization"]
			if len(mcpConnection.HeaderAuth) == 0 || (ok && len(mcpConnection.HeaderAuth) == 1 && strings.HasPrefix(auth, "Bearer ")) {
				externalRequest.AddMcpServerFromMcpConnection(mcpConnection)
				directMcps[mcpConnection.Id] = true
			}
		}
	}

	for _, mcpConnection := range mcpConnections {
		if directMcps[mcpConnection.Id] {
			continue
		}
		tools := mcpToolsByConnectionId[mcpConnection.Id]
		if len(tools) == 0 {
			continue
		}
		externalRequest.AddMcpConnectionAsFunctionCallTools(mcpConnection, tools)
	}

	r.logger.Debug("Request to Anthropic Messages endpoint")

	response, err := r.anthropicMessagesClient.Messages(baseUrl, apiKey, externalRequest)
	if err != nil {
		return nil, fmt.Errorf("failed to call Anthropic Messages API: %w", err)
	}

	contentTypes := lo.Map(response.Content, func(c external_models.ExternalAnthropicMessagesStandardMessageContent, _ int) string {
		return c.Type
	})

	r.logger.Debug("Response from Anthropic Messages endpoint",
		zap.Int("contentItems", len(response.Content)),
		zap.Strings("contentTypes", contentTypes),
	)

	completion := &interfaces.LLMCompletion{
		Usage: &interfaces.LLMUsage{
			InputTokens:  response.Usage.InputTokens,
			OutputTokens: response.Usage.OutputTokens,
		},
	}

	for _, c := range response.Content {
		switch c.Type {
		case "text":
			if c.Text != nil {
				text := *c.Text
				completion.Text = &text
			}
		case "tool_use":
			if c.Id == nil || c.Name == nil {
				return nil, fmt.Errorf("anthropic tool_use missing id or name")
			}
			completion.ToolCalls = append(completion.ToolCalls, interfaces.ToolCall{
				Id:        *c.Id,
				Name:      *c.Name,
				Arguments: c.Input,
			})
		}
	}

	if completion.Text == nil && len(completion.ToolCalls) == 0 {
		r.logger.Warn(
			"No text or tool_use content found in response",
			zap.Stringp("stopReason", response.StopReason),
			zap.Int("contentItems", len(response.Content)),
			zap.Strings("contentTypes", contentTypes),
		)
		return nil, fmt.Errorf("anthropic returned no actionable content: stopReason=%q contentTypes=%v", lo.FromPtr(response.StopReason), contentTypes)
	}

	return completion, nil
}

func (r *AnthropicMessagesStandardChatCompletionRequester) buildSystems(
	messages []*interfaces.ChatCompletionMessage,
) []external_models.ExternalAnthropicMessagesStandardRequestSystem {
	out := make([]external_models.ExternalAnthropicMessagesStandardRequestSystem, 0)
	for _, m := range messages {
		if m.Role != "system" {
			continue
		}
		out = append(out, external_models.ExternalAnthropicMessagesStandardRequestSystem{
			Type: "text",
			Text: m.Content,
		})
	}
	return out
}

// buildRequestMessages translates the domain message stream into anthropic's
// alternating user/assistant shape. role=tool messages become user messages
// containing tool_result content; consecutive tool messages collapse into one.
func (r *AnthropicMessagesStandardChatCompletionRequester) buildRequestMessages(
	messages []*interfaces.ChatCompletionMessage,
) ([]external_models.ExternalAnthropicMessagesStandardMessage, error) {
	out := make([]external_models.ExternalAnthropicMessagesStandardMessage, 0, len(messages))

	flushToolResults := func(buf []external_models.ExternalAnthropicMessagesStandardMessageContent) []external_models.ExternalAnthropicMessagesStandardMessageContent {
		if len(buf) == 0 {
			return buf
		}
		out = append(out, external_models.ExternalAnthropicMessagesStandardMessage{
			Role:    "user",
			Content: buf,
		})
		return nil
	}

	var pendingToolResults []external_models.ExternalAnthropicMessagesStandardMessageContent

	for _, m := range messages {
		switch m.Role {
		case "system":
			continue

		case "user":
			pendingToolResults = flushToolResults(pendingToolResults)
			text := m.Content
			out = append(out, external_models.ExternalAnthropicMessagesStandardMessage{
				Role: "user",
				Content: []external_models.ExternalAnthropicMessagesStandardMessageContent{
					{Type: "text", Text: &text},
				},
			})

		case "assistant":
			pendingToolResults = flushToolResults(pendingToolResults)
			contents := make([]external_models.ExternalAnthropicMessagesStandardMessageContent, 0)
			if m.Content != "" {
				text := m.Content
				contents = append(contents, external_models.ExternalAnthropicMessagesStandardMessageContent{
					Type: "text",
					Text: &text,
				})
			}
			for _, tc := range m.ToolCalls {
				id := tc.Id
				name := tc.Name
				contents = append(contents, external_models.ExternalAnthropicMessagesStandardMessageContent{
					Type:  "tool_use",
					Id:    &id,
					Name:  &name,
					Input: tc.Arguments,
				})
			}
			out = append(out, external_models.ExternalAnthropicMessagesStandardMessage{
				Role:    "assistant",
				Content: contents,
			})

		case "tool":
			callId := m.ToolCallId
			content := m.Content
			pendingToolResults = append(pendingToolResults, external_models.ExternalAnthropicMessagesStandardMessageContent{
				Type:      "tool_result",
				ToolUseId: &callId,
				Content:   &content,
			})

		default:
			return nil, fmt.Errorf("unsupported chat completion message role: %q", m.Role)
		}
	}

	flushToolResults(pendingToolResults)

	return out, nil
}
