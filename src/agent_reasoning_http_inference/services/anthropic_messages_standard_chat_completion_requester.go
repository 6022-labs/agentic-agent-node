package services

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/6022-labs/agentic-agent-node/src/agent_reasoning/models"
	"github.com/6022-labs/agentic-agent-node/src/agent_reasoning/services/interfaces"
	"github.com/6022-labs/agentic-agent-node/src/agent_reasoning/utils"
	"github.com/6022-labs/agentic-agent-node/src/agent_reasoning_http_inference/clients"
	external_models "github.com/6022-labs/agentic-agent-node/src/agent_reasoning_http_inference/models"
	inference_utils "github.com/6022-labs/agentic-agent-node/src/agent_reasoning_http_inference/utils"
	"github.com/google/uuid"
	"github.com/samber/lo"
	"go.uber.org/zap"
)

type AnthropicMessagesStandardChatCompletionRequesterInterface interface {
	ChatCompletion(
		baseUrl, apiKey string,
		hasDirectAccessToMcp bool,
		llmConnection *models.LLMConnection,
		selfTools []*models.Tool,
		mcpTools map[*models.MCPConnection][]*models.Tool,
		messages []*interfaces.ChatCompletionMessage,
	) (*interfaces.LLMCompletion, error)

	Probe(baseUrl, apiKey string, llmConnection *models.LLMConnection) error
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
	selfTools []*models.Tool,
	mcpTools map[*models.MCPConnection][]*models.Tool,
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
	sortedMcpConnections := utils.SortedMcpConnections(mcpTools)
	if hasDirectAccessToMcp {
		// Anthropic only supports MCP servers with no auth or with a single OAuth Bearer Authorization header.
		for _, mcpConnection := range sortedMcpConnections {
			auth, ok := mcpConnection.HeaderAuth["Authorization"]
			if len(mcpConnection.HeaderAuth) == 0 || (ok && len(mcpConnection.HeaderAuth) == 1 && strings.HasPrefix(auth, "Bearer ")) {
				externalRequest.AddMcpServerFromMcpConnection(mcpConnection)
				directMcps[mcpConnection.Id] = true
			}
		}
	}

	for _, mcpConnection := range sortedMcpConnections {
		if directMcps[mcpConnection.Id] {
			continue
		}
		tools := mcpTools[mcpConnection]
		if len(tools) == 0 {
			continue
		}
		externalRequest.AddMcpConnectionAsFunctionCallTools(mcpConnection, tools)
	}

	externalRequest.AddSelfToolsAsFunctionCallTools(selfTools)

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
		ResponseModel: response.Model,
		ResponseId:    response.ID,
	}
	if response.StopReason != nil {
		completion.FinishReasons = []string{*response.StopReason}
	}

	executedByToolUseId := make(map[string]int)
	for _, c := range response.Content {
		switch c.Type {
		case "text":
			if c.Text != nil {
				completion.Texts = append(completion.Texts, *c.Text)
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
		case "mcp_tool_use":
			if c.Id == nil || c.Name == nil {
				return nil, fmt.Errorf("anthropic mcp_tool_use missing id or name")
			}
			completion.ExecutedToolCalls = append(completion.ExecutedToolCalls, interfaces.ExecutedToolCall{
				Id: *c.Id,
				// Align the provider-executed name with the local mcp_<conn>_<tool> convention when the server is one of ours.
				Name:      inference_utils.ProviderExecutedToolFunctionName(mcpTools, lo.FromPtr(c.ServerName), *c.Name),
				Arguments: c.Input,
			})
			executedByToolUseId[*c.Id] = len(completion.ExecutedToolCalls) - 1
		case "mcp_tool_result":
			if c.ToolUseId == nil {
				return nil, fmt.Errorf("anthropic mcp_tool_result missing tool_use_id")
			}
			idx, ok := executedByToolUseId[*c.ToolUseId]
			if !ok {
				return nil, fmt.Errorf("anthropic mcp_tool_result references unknown tool_use_id %q", *c.ToolUseId)
			}
			output, err := json.Marshal(c.Content)
			if err != nil {
				return nil, fmt.Errorf("failed to marshal anthropic mcp_tool_result content: %w", err)
			}
			completion.ExecutedToolCalls[idx].Output = string(output)
			completion.ExecutedToolCalls[idx].IsError = lo.FromPtr(c.IsError)
		}
	}

	if len(completion.Texts) == 0 && len(completion.ToolCalls) == 0 && len(completion.ExecutedToolCalls) == 0 {
		r.logger.Warn(
			"No text or tool_use content found in response",
			zap.Stringp("stopReason", response.StopReason),
			zap.Int("contentItems", len(response.Content)),
			zap.Strings("contentTypes", contentTypes),
		)
	}

	return completion, nil
}

func (r *AnthropicMessagesStandardChatCompletionRequester) Probe(baseUrl, apiKey string, llmConnection *models.LLMConnection) error {
	text := probeMessage
	request := &external_models.ExternalAnthropicMessagesStandardRequest{
		Model:     llmConnection.Model,
		MaxTokens: uint(probeMaxTokens),
		Messages: []external_models.ExternalAnthropicMessagesStandardMessage{{
			Role:    "user",
			Content: []external_models.ExternalAnthropicMessagesStandardMessageContent{{Type: "text", Text: &text}},
		}},
	}

	_, err := r.anthropicMessagesClient.Messages(baseUrl, apiKey, request)
	return err
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
