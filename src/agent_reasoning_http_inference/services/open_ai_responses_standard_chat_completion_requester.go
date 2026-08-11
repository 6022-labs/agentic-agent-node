package services

import (
	"encoding/json"
	"fmt"

	"github.com/6022-labs/agentic-agent-node/src/agent_reasoning/models"
	"github.com/6022-labs/agentic-agent-node/src/agent_reasoning/services/interfaces"
	"github.com/6022-labs/agentic-agent-node/src/agent_reasoning/utils"
	"github.com/6022-labs/agentic-agent-node/src/agent_reasoning_http_inference/clients"
	external_models "github.com/6022-labs/agentic-agent-node/src/agent_reasoning_http_inference/models"
	inference_utils "github.com/6022-labs/agentic-agent-node/src/agent_reasoning_http_inference/utils"
	"github.com/samber/lo"
	"go.uber.org/zap"
)

type OpenAIResponsesStandardChatCompletionRequesterInterface interface {
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

type OpenAIResponsesStandardChatCompletionRequester struct {
	logger                *zap.Logger
	openAIResponsesClient clients.OpenAIResponsesStandardClientInterface
}

func NewOpenAIResponsesStandardChatCompletionRequester(
	logger *zap.Logger,
	openAIResponsesClient clients.OpenAIResponsesStandardClientInterface,
) *OpenAIResponsesStandardChatCompletionRequester {
	return &OpenAIResponsesStandardChatCompletionRequester{
		logger:                logger,
		openAIResponsesClient: openAIResponsesClient,
	}
}

func (r *OpenAIResponsesStandardChatCompletionRequester) ChatCompletion(
	baseUrl, apiKey string,
	hasDirectAccessToMcp bool,
	llmConnection *models.LLMConnection,
	selfTools []*models.Tool,
	mcpTools map[*models.MCPConnection][]*models.Tool,
	messages []*interfaces.ChatCompletionMessage,
) (*interfaces.LLMCompletion, error) {
	inputs, err := r.buildRequestInputs(messages)
	if err != nil {
		return nil, fmt.Errorf("failed to build openai responses inputs: %w", err)
	}

	externalRequest := external_models.NewExternalOpenAIResponsesStandardRequest(llmConnection.Model, inputs, nil)

	for _, mcpConnection := range utils.SortedMcpConnections(mcpTools) {
		tools := mcpTools[mcpConnection]
		if hasDirectAccessToMcp {
			externalRequest.AddMcpConnectionAsMcpTool(mcpConnection, inference_utils.ToolNames(tools))
			continue
		}
		if len(tools) == 0 {
			continue
		}
		externalRequest.AddMcpConnectionAsFunctionCallTools(mcpConnection, tools)
	}

	externalRequest.AddSelfToolsAsFunctionCallTools(selfTools)

	r.logger.Debug("Request to Responses endpoint")

	response, err := r.openAIResponsesClient.Responses(baseUrl, apiKey, externalRequest)
	if err != nil {
		return nil, err
	}

	outputTypes := lo.Map(
		response.Output,
		func(o external_models.ExternalOpenAIResponsesStandardResponseOutput, _ int) string {
			return o.Type
		},
	)

	r.logger.Debug("Response from Responses endpoint",
		zap.Strings("outputTypes", outputTypes),
	)

	completion := &interfaces.LLMCompletion{
		Usage: &interfaces.LLMUsage{
			InputTokens:  response.Usage.InputTokens,
			OutputTokens: response.Usage.OutputTokens,
		},
		ResponseModel: response.Model,
		ResponseId:    response.ID,
	}

	for _, o := range response.Output {
		switch o.Type {
		case "message":
			for _, c := range o.Content {
				if c.Type == "output_text" || c.Type == "text" {
					completion.Texts = append(completion.Texts, c.Text)
				}
			}
		case "function_call":
			if o.Name == nil || o.CallId == nil {
				return nil, fmt.Errorf("openai responses function_call missing name or call_id")
			}
			args, err := o.ParsedArguments()
			if err != nil {
				return nil, fmt.Errorf("failed to parse openai responses function_call arguments: %w", err)
			}
			completion.ToolCalls = append(completion.ToolCalls, interfaces.ToolCall{
				Id:        *o.CallId,
				Name:      *o.Name,
				Arguments: args,
			})
		case "mcp_call":
			if o.Name == nil {
				return nil, fmt.Errorf("openai responses mcp_call missing name")
			}
			args, err := o.ParsedArguments()
			if err != nil {
				return nil, fmt.Errorf("failed to parse openai responses mcp_call arguments: %w", err)
			}
			completion.ExecutedToolCalls = append(completion.ExecutedToolCalls, interfaces.ExecutedToolCall{
				Id: o.ID,
				// Align the provider-executed name with local mcp_<conn>_<tool> when the server is one of ours.
				Name: inference_utils.ProviderExecutedToolFunctionName(
					mcpTools,
					lo.FromPtr(o.ServerLabel),
					*o.Name,
				),
				Arguments: args,
				Output:    lo.FromPtr(o.Output),
				IsError:   o.Error != nil,
			})
		}
	}

	if len(completion.Texts) == 0 && len(completion.ToolCalls) == 0 && len(completion.ExecutedToolCalls) == 0 {
		r.logger.Warn(
			"No actionable output found in response",
			zap.String("status", response.Status),
			zap.Any("incompleteDetails", response.IncompleteDetails),
			zap.Strings("outputTypes", outputTypes),
		)
	}

	return completion, nil
}

func (r *OpenAIResponsesStandardChatCompletionRequester) Probe(
	baseUrl, apiKey string,
	llmConnection *models.LLMConnection,
) error {
	role := "user"
	content := probeMessage
	request := &external_models.ExternalOpenAIResponsesStandardRequest{
		Model: llmConnection.Model,
		Input: []external_models.ExternalOpenAIResponsesStandardRequestInput{
			{Type: "message", Role: &role, Content: &content},
		},
		Tools:           []external_models.ExternalOpenAIResponsesStandardRequestTool{},
		MaxOutputTokens: lo.ToPtr(probeMaxTokens),
	}

	_, err := r.openAIResponsesClient.Responses(baseUrl, apiKey, request)
	return err
}

// buildRequestInputs flattens domain messages into the OpenAI Responses input
// stream: assistant tool_calls split into a text message + one function_call
// each; role=tool messages become function_call_output entries.
func (r *OpenAIResponsesStandardChatCompletionRequester) buildRequestInputs(
	messages []*interfaces.ChatCompletionMessage,
) ([]external_models.ExternalOpenAIResponsesStandardRequestInput, error) {
	out := make([]external_models.ExternalOpenAIResponsesStandardRequestInput, 0, len(messages))

	for _, m := range messages {
		switch m.Role {
		case "system", "user":
			role := m.Role
			content := m.Content
			out = append(out, external_models.ExternalOpenAIResponsesStandardRequestInput{
				Type:    "message",
				Role:    &role,
				Content: &content,
			})

		case "assistant":
			if m.Content != "" {
				role := "assistant"
				content := m.Content
				out = append(out, external_models.ExternalOpenAIResponsesStandardRequestInput{
					Type:    "message",
					Role:    &role,
					Content: &content,
				})
			}
			for _, tc := range m.ToolCalls {
				argsBytes, err := json.Marshal(tc.Arguments)
				if err != nil {
					return nil, fmt.Errorf("failed to marshal tool call arguments: %w", err)
				}
				argsStr := string(argsBytes)
				name := tc.Name
				callId := tc.Id
				out = append(out, external_models.ExternalOpenAIResponsesStandardRequestInput{
					Type:      "function_call",
					Name:      &name,
					CallId:    &callId,
					Arguments: &argsStr,
				})
			}

		case "tool":
			callId := m.ToolCallId
			content := m.Content
			out = append(out, external_models.ExternalOpenAIResponsesStandardRequestInput{
				Type:   "function_call_output",
				CallId: &callId,
				Output: &content,
			})

		default:
			return nil, fmt.Errorf("unsupported chat completion message role: %q", m.Role)
		}
	}

	return out, nil
}
