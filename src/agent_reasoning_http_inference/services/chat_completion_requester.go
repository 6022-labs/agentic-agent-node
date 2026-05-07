package services

import (
	"context"
	"fmt"

	"github.com/6022-labs/agentic-agent-node/src/agent_reasoning/models"
	"github.com/6022-labs/agentic-agent-node/src/agent_reasoning/models/enums"
	"github.com/6022-labs/agentic-agent-node/src/agent_reasoning/services/interfaces"
	reasoning_interfaces "github.com/6022-labs/agentic-agent-node/src/agent_reasoning/services/interfaces"
	"github.com/6022-labs/agentic-agent-node/src/agent_reasoning_http_inference/settings"
	"github.com/google/uuid"
	"github.com/pkoukk/tiktoken-go"
	"go.uber.org/zap"
)

// ChatCompletionRequester is the http-side adapter that implements
// ChatCompletionRequesterInterface. It picks the right per-platform requester
// based on LLMConnection.Platform, fetches the api key from Vault, trims the
// message stream to the configured token budget, and delegates the round-trip.
//
// It does NOT loop on tool calls — tool resolution is the inference_loop's job.
type ChatCompletionRequester struct {
	logger *zap.Logger

	tokenSettings    *settings.TokenSettings
	llmSecretStorage reasoning_interfaces.LLMSecretStorageInterface

	xSettings             *settings.XSettings
	openAiSettings        *settings.OpenAISettings
	ollamaSettings        *settings.OllamaSettings
	mistralSettings       *settings.MistralSettings
	deepSeekSettings      *settings.DeepSeekSettings
	libertAiSettings      *settings.LibertAISettings
	anthropicSettings     *settings.AnthropicSettings
	amazonBedrockSettings *settings.AmazonBedrockSettings

	openAIResponsesStandardChatCompletionRequester       OpenAIResponsesStandardChatCompletionRequesterInterface
	ollamaChatMessageStandardChatCompletionRequester     OllamaChatMessageStandardChatCompletionRequesterInterface
	anthropicMessagesStandardChatCompletionRequester     AnthropicMessagesStandardChatCompletionRequesterInterface
	openAICompletionsStandardChatCompletionRequester     OpenAICompletionsStandardChatCompletionRequesterInterface
	amazonBedrockConverseStandardChatCompletionRequester AmazonBedrockConverseStandardChatCompletionRequesterInterface
}

func NewChatCompletionRequester(
	logger *zap.Logger,
	xSettings *settings.XSettings,
	tokenSettings *settings.TokenSettings,
	llmSecretStorage reasoning_interfaces.LLMSecretStorageInterface,
	openAiSettings *settings.OpenAISettings,
	ollamaSettings *settings.OllamaSettings,
	mistralSettings *settings.MistralSettings,
	deepSeekSettings *settings.DeepSeekSettings,
	libertAiSettings *settings.LibertAISettings,
	anthropicSettings *settings.AnthropicSettings,
	amazonBedrockSettings *settings.AmazonBedrockSettings,
	openAIResponsesStandardChatCompletionRequester OpenAIResponsesStandardChatCompletionRequesterInterface,
	ollamaChatMessageStandardChatCompletionRequester OllamaChatMessageStandardChatCompletionRequesterInterface,
	anthropicMessagesStandardChatCompletionRequester AnthropicMessagesStandardChatCompletionRequesterInterface,
	openAICompletionsStandardChatCompletionRequester OpenAICompletionsStandardChatCompletionRequesterInterface,
	amazonBedrockConverseStandardChatCompletionRequester AmazonBedrockConverseStandardChatCompletionRequesterInterface,
) *ChatCompletionRequester {
	return &ChatCompletionRequester{
		logger: logger,

		tokenSettings:    tokenSettings,
		llmSecretStorage: llmSecretStorage,

		xSettings:             xSettings,
		openAiSettings:        openAiSettings,
		ollamaSettings:        ollamaSettings,
		mistralSettings:       mistralSettings,
		deepSeekSettings:      deepSeekSettings,
		libertAiSettings:      libertAiSettings,
		anthropicSettings:     anthropicSettings,
		amazonBedrockSettings: amazonBedrockSettings,

		openAIResponsesStandardChatCompletionRequester:       openAIResponsesStandardChatCompletionRequester,
		ollamaChatMessageStandardChatCompletionRequester:     ollamaChatMessageStandardChatCompletionRequester,
		openAICompletionsStandardChatCompletionRequester:     openAICompletionsStandardChatCompletionRequester,
		anthropicMessagesStandardChatCompletionRequester:     anthropicMessagesStandardChatCompletionRequester,
		amazonBedrockConverseStandardChatCompletionRequester: amazonBedrockConverseStandardChatCompletionRequester,
	}
}

func (r *ChatCompletionRequester) ChatCompletion(
	ctx context.Context,
	llmConnection *models.LLMConnection,
	mcpConnections []*models.MCPConnection,
	mcpToolsByConnectionId map[uuid.UUID][]*reasoning_interfaces.McpTool,
	messages []*interfaces.ChatCompletionMessage,
) (*interfaces.LLMCompletion, error) {
	trimmedMessages, err := r.trimMessagesToTokenLimit(messages, r.tokenSettings.MaxTokensPerRequest)
	if err != nil {
		return nil, err
	}

	platformSettings, err := r.getPlatformSettings(llmConnection.Platform)
	if err != nil {
		return nil, err
	}

	apiKey, err := r.resolveApiKey(ctx)
	if err != nil {
		return nil, err
	}

	return r.completeUsingAppropriateStandard(
		llmConnection,
		mcpConnections,
		mcpToolsByConnectionId,
		trimmedMessages,
		platformSettings,
		apiKey,
	)
}

func (r *ChatCompletionRequester) resolveApiKey(ctx context.Context) (*string, error) {
	apiKey, found, err := r.llmSecretStorage.Get(ctx)
	if err != nil {
		return nil, err
	}
	if !found {
		return nil, nil
	}
	return &apiKey, nil
}

// trimMessagesToTokenLimit trims messages from the start until the total token
// count is within the limit. System messages stay at the head; the trimmer
// drops the oldest non-system messages first. Token budget is approximated
// using OpenAI's gpt-4 tokenizer for all platforms — close enough for trim.
func (r *ChatCompletionRequester) trimMessagesToTokenLimit(
	messages []*interfaces.ChatCompletionMessage,
	limit uint,
) ([]*interfaces.ChatCompletionMessage, error) {
	tokenizer, err := tiktoken.EncodingForModel("gpt-4")
	if err != nil {
		return nil, fmt.Errorf("failed to get tokenizer: %w", err)
	}

	totalTokens := uint(0)
	msgTokens := make([]uint, len(messages))
	for i, msg := range messages {
		tokens := tokenizer.Encode(msg.Content, nil, nil)
		msgTokens[i] = uint(len(tokens))
		totalTokens += uint(len(tokens))
	}

	startIdx := 0
	for totalTokens > limit && startIdx < len(messages) {
		totalTokens -= msgTokens[startIdx]
		startIdx++
	}

	return messages[startIdx:], nil
}

func (r *ChatCompletionRequester) getPlatformSettings(platform enums.LLMConnectionPlatform) (settings.StandardSettingsInterface, error) {
	switch platform {
	case enums.LLMConnectionPlatformLibertAI:
		return r.libertAiSettings, nil
	case enums.LLMConnectionPlatformOpenAI:
		return r.openAiSettings, nil
	case enums.LLMConnectionPlatformOllama:
		return r.ollamaSettings, nil
	case enums.LLMConnectionPlatformMistral:
		return r.mistralSettings, nil
	case enums.LLMConnectionPlatformDeepseek:
		return r.deepSeekSettings, nil
	case enums.LLMConnectionPlatformX:
		return r.xSettings, nil
	case enums.LLMConnectionPlatformAnthropic:
		return r.anthropicSettings, nil
	case enums.LLMConnectionPlatformAmazonBedrock:
		return r.amazonBedrockSettings, nil
	default:
		return nil, fmt.Errorf("unsupported platform: %s", platform)
	}
}

func (r *ChatCompletionRequester) completeUsingAppropriateStandard(
	llmConnection *models.LLMConnection,
	mcpConnections []*models.MCPConnection,
	mcpToolsByConnectionId map[uuid.UUID][]*reasoning_interfaces.McpTool,
	messages []*interfaces.ChatCompletionMessage,
	platformSettings settings.StandardSettingsInterface,
	apiKey *string,
) (*interfaces.LLMCompletion, error) {
	switch platformSettings.PlatformAPI() {
	case settings.PlatformAPIOpenAIResponsesStandard:
		if apiKey == nil {
			return nil, fmt.Errorf("api key is required for OpenAI Responses Standard")
		}
		return r.openAIResponsesStandardChatCompletionRequester.ChatCompletion(
			platformSettings.GetBaseUrl(),
			*apiKey,
			platformSettings.HasDirectAccessToMcp(),
			llmConnection,
			mcpConnections,
			mcpToolsByConnectionId,
			messages,
		)

	case settings.PlatformAPIOpenAICompletionsStandard:
		if apiKey == nil {
			return nil, fmt.Errorf("api key is required for OpenAI Completions Standard")
		}
		return r.openAICompletionsStandardChatCompletionRequester.ChatCompletion(
			platformSettings.GetBaseUrl(),
			*apiKey,
			llmConnection,
			mcpConnections,
			mcpToolsByConnectionId,
			messages,
		)

	case settings.PlatformAPIOllamaChatMessageStandard:
		if llmConnection.BaseUrl == nil {
			return nil, fmt.Errorf("base URL is required for Ollama Chat Message Standard")
		}
		return r.ollamaChatMessageStandardChatCompletionRequester.ChatCompletion(
			*llmConnection.BaseUrl,
			llmConnection,
			mcpConnections,
			mcpToolsByConnectionId,
			messages,
		)

	case settings.PlatformAPIAnthropicMessagesStandard:
		if apiKey == nil {
			return nil, fmt.Errorf("api key is required for Anthropic Messages Standard")
		}
		return r.anthropicMessagesStandardChatCompletionRequester.ChatCompletion(
			platformSettings.GetBaseUrl(),
			*apiKey,
			platformSettings.HasDirectAccessToMcp(),
			llmConnection,
			mcpConnections,
			mcpToolsByConnectionId,
			messages,
		)

	case settings.PlatformAPIAmazonBedrockConverseStandard:
		if apiKey == nil {
			return nil, fmt.Errorf("api key is required for Amazon Bedrock Converse Standard")
		}
		if llmConnection.BaseUrl == nil {
			return nil, fmt.Errorf("base URL is required for Amazon Bedrock Converse Standard")
		}
		return r.amazonBedrockConverseStandardChatCompletionRequester.ChatCompletion(
			*llmConnection.BaseUrl,
			*apiKey,
			llmConnection,
			mcpConnections,
			mcpToolsByConnectionId,
			messages,
		)

	default:
		return nil, fmt.Errorf("unsupported platform API: %s", platformSettings.PlatformAPI())
	}
}
