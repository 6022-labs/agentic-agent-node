package services

import (
	"fmt"

	"github.com/6022-labs/agentic-agent-node/src/agent_reasoning/models"
	"github.com/6022-labs/agentic-agent-node/src/agent_reasoning/services/interfaces"
	"github.com/6022-labs/agentic-agent-node/src/agent_reasoning/utils"
	"github.com/6022-labs/agentic-agent-node/src/agent_reasoning_http_inference/clients"
	external_models "github.com/6022-labs/agentic-agent-node/src/agent_reasoning_http_inference/models"
	"github.com/samber/lo"
	"go.uber.org/zap"
)

type AmazonBedrockConverseStandardChatCompletionRequesterInterface interface {
	ChatCompletion(
		baseUrl, apiKey string,
		llmConnection *models.LLMConnection,
		selfTools []*models.Tool,
		mcpTools map[*models.MCPConnection][]*models.Tool,
		messages []*interfaces.ChatCompletionMessage,
	) (*interfaces.LLMCompletion, error)

	Probe(baseUrl, apiKey string, llmConnection *models.LLMConnection) error
}

type AmazonBedrockConverseStandardChatCompletionRequester struct {
	logger                      *zap.Logger
	amazonBedrockConverseClient clients.AmazonBedrockConverseStandardClientInterface
}

func NewAmazonBedrockConverseStandardChatCompletionRequester(
	logger *zap.Logger,
	amazonBedrockConverseClient clients.AmazonBedrockConverseStandardClientInterface,
) *AmazonBedrockConverseStandardChatCompletionRequester {
	return &AmazonBedrockConverseStandardChatCompletionRequester{
		logger:                      logger,
		amazonBedrockConverseClient: amazonBedrockConverseClient,
	}
}

func (r *AmazonBedrockConverseStandardChatCompletionRequester) ChatCompletion(
	baseUrl, apiKey string,
	llmConnection *models.LLMConnection,
	selfTools []*models.Tool,
	mcpTools map[*models.MCPConnection][]*models.Tool,
	messages []*interfaces.ChatCompletionMessage,
) (*interfaces.LLMCompletion, error) {
	systems := r.buildSystems(messages)
	extMessages, err := r.buildRequestMessages(messages)
	if err != nil {
		return nil, fmt.Errorf("failed to build bedrock messages: %w", err)
	}

	externalRequest := external_models.NewExternalAmazonBedrockConverseStandardRequest(llmConnection.Model, extMessages, systems)

	for _, mcpConnection := range utils.SortedMcpConnections(mcpTools) {
		tools := mcpTools[mcpConnection]
		if len(tools) == 0 {
			continue
		}
		externalRequest.AddMcpConnectionAsFunctionCallTools(mcpConnection, tools)
	}

	externalRequest.AddSelfToolsAsFunctionCallTools(selfTools)

	r.logger.Debug("Request to Amazon Bedrock Converse Messages endpoint")

	response, err := r.amazonBedrockConverseClient.Messages(baseUrl, apiKey, externalRequest)
	if err != nil {
		return nil, fmt.Errorf("failed to call Amazon Bedrock Converse API: %w", err)
	}

	contentTypes := lo.Map(response.Output.Message.Content, func(c external_models.ExternalAmazonBedrockConverseStandardMessageContent, _ int) string {
		return c.ContentType()
	})

	r.logger.Debug("Response from Amazon Bedrock Converse Messages endpoint",
		zap.Strings("contentTypes", contentTypes),
	)

	completion := &interfaces.LLMCompletion{
		Usage: &interfaces.LLMUsage{
			InputTokens:  response.Usage.InputTokens,
			OutputTokens: response.Usage.OutputTokens,
		},
		// Converse doesn't echo the model; llmConnection.Model is what served the response.
		ResponseModel: llmConnection.Model,
	}
	if response.StopReason != nil {
		completion.FinishReasons = []string{*response.StopReason}
	}

	for _, c := range response.Output.Message.Content {
		switch {
		case c.Text != nil:
			completion.Texts = append(completion.Texts, *c.Text)
		case c.ToolUse != nil:
			completion.ToolCalls = append(completion.ToolCalls, interfaces.ToolCall{
				Id:        c.ToolUse.ToolUseId,
				Name:      c.ToolUse.Name,
				Arguments: c.ToolUse.Input,
			})
		}
	}

	if len(completion.Texts) == 0 && len(completion.ToolCalls) == 0 {
		r.logger.Warn(
			"No text or tool_use content found in response",
			zap.Stringp("stopReason", response.StopReason),
			zap.Int("contentItems", len(response.Output.Message.Content)),
			zap.Strings("contentTypes", contentTypes),
		)
	}

	return completion, nil
}

func (r *AmazonBedrockConverseStandardChatCompletionRequester) Probe(baseUrl, apiKey string, llmConnection *models.LLMConnection) error {
	text := probeMessage
	request := &external_models.ExternalAmazonBedrockConverseStandardRequest{
		ModelId: llmConnection.Model,
		System:  []external_models.ExternalAmazonBedrockConverseStandardRequestSystem{},
		Messages: []external_models.ExternalAmazonBedrockConverseStandardMessage{{
			Role:    "user",
			Content: []external_models.ExternalAmazonBedrockConverseStandardMessageContent{{Text: &text}},
		}},
		InferenceConfig: &external_models.ExternalAmazonBedrockConverseStandardRequestInferenceConfig{MaxTokens: probeMaxTokens},
	}

	_, err := r.amazonBedrockConverseClient.Messages(baseUrl, apiKey, request)
	return err
}

func (r *AmazonBedrockConverseStandardChatCompletionRequester) buildSystems(
	messages []*interfaces.ChatCompletionMessage,
) []external_models.ExternalAmazonBedrockConverseStandardRequestSystem {
	out := make([]external_models.ExternalAmazonBedrockConverseStandardRequestSystem, 0)
	for _, m := range messages {
		if m.Role != "system" {
			continue
		}
		out = append(out, external_models.ExternalAmazonBedrockConverseStandardRequestSystem{Text: m.Content})
	}
	return out
}

// buildRequestMessages translates domain messages into bedrock's
// alternating user/assistant shape. role=tool messages collapse into a single
// user message containing toolResult content entries.
func (r *AmazonBedrockConverseStandardChatCompletionRequester) buildRequestMessages(
	messages []*interfaces.ChatCompletionMessage,
) ([]external_models.ExternalAmazonBedrockConverseStandardMessage, error) {
	out := make([]external_models.ExternalAmazonBedrockConverseStandardMessage, 0, len(messages))

	flushToolResults := func(buf []external_models.ExternalAmazonBedrockConverseStandardMessageContent) []external_models.ExternalAmazonBedrockConverseStandardMessageContent {
		if len(buf) == 0 {
			return buf
		}
		out = append(out, external_models.ExternalAmazonBedrockConverseStandardMessage{
			Role:    "user",
			Content: buf,
		})
		return nil
	}

	var pendingToolResults []external_models.ExternalAmazonBedrockConverseStandardMessageContent

	for _, m := range messages {
		switch m.Role {
		case "system":
			continue

		case "user":
			pendingToolResults = flushToolResults(pendingToolResults)
			text := m.Content
			out = append(out, external_models.ExternalAmazonBedrockConverseStandardMessage{
				Role:    "user",
				Content: []external_models.ExternalAmazonBedrockConverseStandardMessageContent{{Text: &text}},
			})

		case "assistant":
			pendingToolResults = flushToolResults(pendingToolResults)
			contents := make([]external_models.ExternalAmazonBedrockConverseStandardMessageContent, 0)
			if m.Content != "" {
				text := m.Content
				contents = append(contents, external_models.ExternalAmazonBedrockConverseStandardMessageContent{Text: &text})
			}
			for _, tc := range m.ToolCalls {
				contents = append(contents, external_models.ExternalAmazonBedrockConverseStandardMessageContent{
					ToolUse: &external_models.ExternalAmazonBedrockConverseStandardMessageContentToolUse{
						Name:      tc.Name,
						Input:     tc.Arguments,
						ToolUseId: tc.Id,
					},
				})
			}
			out = append(out, external_models.ExternalAmazonBedrockConverseStandardMessage{
				Role:    "assistant",
				Content: contents,
			})

		case "tool":
			content := m.Content
			pendingToolResults = append(pendingToolResults, external_models.ExternalAmazonBedrockConverseStandardMessageContent{
				ToolResult: &external_models.ExternalAmazonBedrockConverseStandardMessageContentToolResult{
					ToolUseId: m.ToolCallId,
					Content: []external_models.ExternalAmazonBedrockConverseStandardMessageContentToolResultContent{
						{Text: &content},
					},
				},
			})

		default:
			return nil, fmt.Errorf("unsupported chat completion message role: %q", m.Role)
		}
	}

	flushToolResults(pendingToolResults)

	return out, nil
}
