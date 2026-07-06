package services

import (
	"encoding/json"
	"fmt"

	"github.com/6022-labs/agentic-agent-node/src/agent_reasoning/models"
	"github.com/6022-labs/agentic-agent-node/src/agent_reasoning/services/interfaces"
	"github.com/6022-labs/agentic-agent-node/src/agent_reasoning/utils"
	"github.com/6022-labs/agentic-agent-node/src/agent_reasoning_http_inference/clients"
	external_models "github.com/6022-labs/agentic-agent-node/src/agent_reasoning_http_inference/models"
	"go.uber.org/zap"
)

type OpenAICompletionsStandardChatCompletionRequesterInterface interface {
	ChatCompletion(
		baseUrl, apiKey string,
		llmConnection *models.LLMConnection,
		selfTools []*models.Tool,
		mcpTools map[*models.MCPConnection][]*models.Tool,
		messages []*interfaces.ChatCompletionMessage,
	) (*interfaces.LLMCompletion, error)
}

type OpenAICompletionsStandardChatCompletionRequester struct {
	logger                  *zap.Logger
	openAICompletionsClient clients.OpenAICompletionsStandardClientInterface
}

func NewOpenAICompletionsStandardChatCompletionRequester(
	logger *zap.Logger,
	openAICompletionsClient clients.OpenAICompletionsStandardClientInterface,
) *OpenAICompletionsStandardChatCompletionRequester {
	return &OpenAICompletionsStandardChatCompletionRequester{
		logger:                  logger,
		openAICompletionsClient: openAICompletionsClient,
	}
}

func (r *OpenAICompletionsStandardChatCompletionRequester) ChatCompletion(
	baseUrl, apiKey string,
	llmConnection *models.LLMConnection,
	selfTools []*models.Tool,
	mcpTools map[*models.MCPConnection][]*models.Tool,
	messages []*interfaces.ChatCompletionMessage,
) (*interfaces.LLMCompletion, error) {
	extMessages, err := r.buildRequestMessages(messages)
	if err != nil {
		return nil, fmt.Errorf("failed to build openai completions messages: %w", err)
	}

	externalRequest := external_models.NewExternalOpenAICompletionsStandardRequest(llmConnection.Model, extMessages)

	for _, mcpConnection := range utils.SortedMcpConnections(mcpTools) {
		tools := mcpTools[mcpConnection]
		if len(tools) == 0 {
			continue
		}
		externalRequest.AddMcpConnectionAsFunctionCallTools(mcpConnection, tools)
	}

	externalRequest.AddSelfToolsAsFunctionCallTools(selfTools)

	r.logger.Debug("Request to Completions endpoint")

	response, err := r.openAICompletionsClient.ChatCompletion(baseUrl, apiKey, externalRequest)
	if err != nil {
		return nil, err
	}

	if len(response.Choices) == 0 {
		r.logger.Warn(
			"No choices found in response",
			zap.String("responseId", response.ID),
			zap.String("model", response.Model),
		)

		return nil, fmt.Errorf("openai completions returned no choices: id=%s model=%s", response.ID, response.Model)
	}

	choice := response.Choices[0]
	completion := &interfaces.LLMCompletion{
		Usage: &interfaces.LLMUsage{
			InputTokens:  response.Usage.PromptTokens,
			OutputTokens: response.Usage.CompletionTokens,
		},
		ResponseModel: response.Model,
		ResponseId:    response.ID,
	}
	if choice.FinishReason != "" {
		completion.FinishReasons = []string{choice.FinishReason}
	}

	if choice.Message.Content != "" {
		text := choice.Message.Content
		completion.Text = &text
	}

	for _, tc := range choice.Message.ToolCalls {
		args, err := tc.Function.ParsedArguments()
		if err != nil {
			return nil, fmt.Errorf("failed to parse openai completions tool call arguments: %w", err)
		}
		completion.ToolCalls = append(completion.ToolCalls, interfaces.ToolCall{
			Id:        tc.ID,
			Name:      tc.Function.Name,
			Arguments: args,
		})
	}

	return completion, nil
}

func (r *OpenAICompletionsStandardChatCompletionRequester) buildRequestMessages(
	messages []*interfaces.ChatCompletionMessage,
) ([]external_models.ExternalOpenAICompletionsStandardRequestMessage, error) {
	out := make([]external_models.ExternalOpenAICompletionsStandardRequestMessage, 0, len(messages))

	for _, m := range messages {
		switch m.Role {
		case "system", "user":
			content := m.Content
			out = append(out, external_models.ExternalOpenAICompletionsStandardRequestMessage{
				Role:    m.Role,
				Content: &content,
			})

		case "assistant":
			msg := external_models.ExternalOpenAICompletionsStandardRequestMessage{Role: "assistant"}
			if m.Content != "" {
				content := m.Content
				msg.Content = &content
			}
			for _, tc := range m.ToolCalls {
				argsBytes, err := json.Marshal(tc.Arguments)
				if err != nil {
					return nil, fmt.Errorf("failed to marshal tool call arguments: %w", err)
				}
				msg.ToolCalls = append(msg.ToolCalls, external_models.ExternalOpenAICompletionsStandardRequestMessageToolCall{
					ID:   tc.Id,
					Type: "function",
					Function: external_models.ExternalOpenAICompletionsStandardToolCallFunction{
						Name:      tc.Name,
						Arguments: string(argsBytes),
					},
				})
			}
			out = append(out, msg)

		case "tool":
			callId := m.ToolCallId
			content := m.Content
			out = append(out, external_models.ExternalOpenAICompletionsStandardRequestMessage{
				Role:       "tool",
				ToolCallId: &callId,
				Content:    &content,
			})

		default:
			return nil, fmt.Errorf("unsupported chat completion message role: %q", m.Role)
		}
	}

	return out, nil
}
