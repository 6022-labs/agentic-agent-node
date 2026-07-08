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

type OllamaChatMessageStandardChatCompletionRequesterInterface interface {
	ChatCompletion(
		baseUrl string,
		llmConnection *models.LLMConnection,
		selfTools []*models.Tool,
		mcpTools map[*models.MCPConnection][]*models.Tool,
		messages []*interfaces.ChatCompletionMessage,
	) (*interfaces.LLMCompletion, error)
}

type OllamaChatMessageStandardChatCompletionRequester struct {
	logger                  *zap.Logger
	ollamaChatMessageClient clients.OllamaChatMessageStandardClientInterface
}

func NewOllamaChatMessageStandardChatCompletionRequester(
	logger *zap.Logger,
	ollamaChatMessageClient clients.OllamaChatMessageStandardClientInterface,
) *OllamaChatMessageStandardChatCompletionRequester {
	return &OllamaChatMessageStandardChatCompletionRequester{
		logger:                  logger,
		ollamaChatMessageClient: ollamaChatMessageClient,
	}
}

func (r *OllamaChatMessageStandardChatCompletionRequester) ChatCompletion(
	baseUrl string,
	llmConnection *models.LLMConnection,
	selfTools []*models.Tool,
	mcpTools map[*models.MCPConnection][]*models.Tool,
	messages []*interfaces.ChatCompletionMessage,
) (*interfaces.LLMCompletion, error) {
	extMessages, err := r.buildRequestMessages(messages)
	if err != nil {
		return nil, fmt.Errorf("failed to build ollama messages: %w", err)
	}

	externalRequest := external_models.NewExternalOllamaChatMessageStandardRequest(llmConnection.Model, extMessages)

	for _, mcpConnection := range utils.SortedMcpConnections(mcpTools) {
		tools := mcpTools[mcpConnection]
		if len(tools) == 0 {
			continue
		}
		externalRequest.AddMcpConnectionAsFunctionCallTools(mcpConnection, tools)
	}

	externalRequest.AddSelfToolsAsFunctionCallTools(selfTools)

	r.logger.Debug("Request to Chat Message endpoint")

	response, err := r.ollamaChatMessageClient.ChatMessage(baseUrl, externalRequest)
	if err != nil {
		return nil, err
	}

	r.logger.Debug("Response from Chat Message endpoint",
		zap.Int("toolCallCount", len(response.Message.ToolCalls)),
	)

	completion := &interfaces.LLMCompletion{
		Usage: &interfaces.LLMUsage{
			InputTokens:  response.PromptEvalCount,
			OutputTokens: response.EvalCount,
		},
		ResponseModel: response.Model,
	}
	if response.DoneReason != nil {
		completion.FinishReasons = []string{*response.DoneReason}
	}

	if response.Message.Content != "" {
		completion.Texts = append(completion.Texts, response.Message.Content)
	}

	for _, tc := range response.Message.ToolCalls {
		if tc.Function == nil {
			continue
		}
		// Ollama tool calls have no id; use the function name as a stable id.
		// The inner loop echoes it back in the tool message's ToolCallId.
		completion.ToolCalls = append(completion.ToolCalls, interfaces.ToolCall{
			Id:        tc.Function.Name,
			Name:      tc.Function.Name,
			Arguments: tc.Function.Arguments,
		})
	}

	if len(completion.Texts) == 0 && len(completion.ToolCalls) == 0 {
		r.logger.Warn(
			"Empty content in response message",
			zap.Stringp("doneReason", response.DoneReason),
		)
		return nil, fmt.Errorf("ollama returned empty content: doneReason=%q", lo.FromPtr(response.DoneReason))
	}

	return completion, nil
}

func (r *OllamaChatMessageStandardChatCompletionRequester) buildRequestMessages(
	messages []*interfaces.ChatCompletionMessage,
) ([]external_models.ExternalOllamaChatMessageStandardMessage, error) {
	out := make([]external_models.ExternalOllamaChatMessageStandardMessage, 0, len(messages))

	for _, m := range messages {
		switch m.Role {
		case "system", "user":
			out = append(out, external_models.ExternalOllamaChatMessageStandardMessage{
				Role:    m.Role,
				Content: m.Content,
			})

		case "assistant":
			extMsg := external_models.ExternalOllamaChatMessageStandardMessage{
				Role:    "assistant",
				Content: m.Content,
			}
			for _, tc := range m.ToolCalls {
				name := tc.Name
				extMsg.ToolCalls = append(extMsg.ToolCalls, external_models.ExternalOllamaChatMessageStandardMessageToolCall{
					Function: &external_models.ExternalOllamaChatMessageStandardMessageToolCallFunction{
						Name:      name,
						Arguments: tc.Arguments,
					},
				})
			}
			out = append(out, extMsg)

		case "tool":
			toolName := m.ToolCallId
			out = append(out, external_models.ExternalOllamaChatMessageStandardMessage{
				Role:     "tool",
				ToolName: &toolName,
				Content:  m.Content,
			})

		default:
			return nil, fmt.Errorf("unsupported chat completion message role: %q", m.Role)
		}
	}

	return out, nil
}
