package services

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/6022-labs/agentic-agent-node/src/agent_reasoning/models"
	"github.com/6022-labs/agentic-agent-node/src/agent_reasoning/models/enums"
	"github.com/6022-labs/agentic-agent-node/src/agent_reasoning/services/interfaces"
	reasoning_interfaces "github.com/6022-labs/agentic-agent-node/src/agent_reasoning/services/interfaces"
	metrics_interfaces "github.com/6022-labs/agentic-agent-node/src/agent_reasoning_http_inference/metrics/interfaces"
	"github.com/6022-labs/agentic-agent-node/src/agent_reasoning_http_inference/settings"
	"github.com/pkoukk/tiktoken-go"
	"go.uber.org/zap"
)

// Probe request shape: 16 is the smallest cap OpenAI Responses accepts.
const (
	probeMessage   = "ping"
	probeMaxTokens = 16
)

// ChatCompletionRequester picks the per-platform requester by
// LLMConnection.Platform, fetches the api key from Vault, trims messages to the
// token budget, and delegates. Does NOT loop on tool calls — that's inference_loop.
type ChatCompletionRequester struct {
	logger *zap.Logger

	tokenSettings    *settings.TokenSettings
	llmSecretStorage reasoning_interfaces.LLMSecretStorageInterface
	metrics          metrics_interfaces.LlmInferenceMetricsInterface

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

	tokenizerOnce sync.Once
	tokenizer     *tiktoken.Tiktoken
	tokenizerErr  error
}

func NewChatCompletionRequester(
	logger *zap.Logger,
	xSettings *settings.XSettings,
	tokenSettings *settings.TokenSettings,
	llmSecretStorage reasoning_interfaces.LLMSecretStorageInterface,
	metrics metrics_interfaces.LlmInferenceMetricsInterface,
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
		metrics:          metrics,

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
	selfTools []*models.Tool,
	mcpTools map[*models.MCPConnection][]*models.Tool,
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

	baseUrl, err := r.resolveBaseUrl(llmConnection, platformSettings)
	if err != nil {
		return nil, err
	}

	call := metrics_interfaces.CompletionCall{
		Provider:     string(llmConnection.Platform),
		RequestModel: llmConnection.Model,
		ServerUrl:    baseUrl,
	}

	start := time.Now()
	completion, err := r.completeUsingAppropriateStandard(
		baseUrl,
		llmConnection,
		selfTools,
		mcpTools,
		trimmedMessages,
		platformSettings,
		apiKey,
	)
	duration := time.Since(start)

	// Per GenAI semconv, request volume and error rate come from the duration
	// histogram's count + the error.type attribute (no separate counters).
	if err != nil {
		r.metrics.RecordOperationDuration(ctx, call, duration, err)
		return nil, err
	}

	// Providers may answer from a different model than requested (routing, aliases, fallbacks).
	if completion != nil {
		call.ResponseModel = completion.ResponseModel
	}
	r.metrics.RecordOperationDuration(ctx, call, duration, nil)

	if completion != nil && completion.Usage != nil {
		r.metrics.RecordTokenUsage(ctx, call, int64(completion.Usage.InputTokens), int64(completion.Usage.OutputTokens))
	}

	return completion, nil
}

func (r *ChatCompletionRequester) ProbeConnection(
	ctx context.Context,
	llmConnection *models.LLMConnection,
	apiKey string,
) error {
	platformSettings, err := r.getPlatformSettings(llmConnection.Platform)
	if err != nil {
		return err
	}

	key := &apiKey
	if apiKey == "" {
		key, err = r.resolveApiKey(ctx)
		if err != nil {
			return err
		}
	}

	baseUrl, err := r.resolveBaseUrl(llmConnection, platformSettings)
	if err != nil {
		return err
	}

	start := time.Now()
	err = r.probeUsingAppropriateStandard(baseUrl, llmConnection, platformSettings, key)

	call := metrics_interfaces.CompletionCall{
		Provider:     string(llmConnection.Platform),
		RequestModel: llmConnection.Model,
		ServerUrl:    baseUrl,
	}
	r.metrics.RecordProbe(ctx, call, time.Since(start), err)

	return err
}

// A platform leaving GetBaseUrl empty is self-hosted: its url lives on the connection instead.
func (r *ChatCompletionRequester) resolveBaseUrl(
	llmConnection *models.LLMConnection,
	platformSettings settings.StandardSettingsInterface,
) (string, error) {
	if baseUrl := platformSettings.GetBaseUrl(); baseUrl != "" {
		return baseUrl, nil
	}
	if llmConnection.BaseUrl == nil {
		return "", fmt.Errorf("base URL is required for %s", platformSettings.PlatformAPI())
	}
	return *llmConnection.BaseUrl, nil
}

func (r *ChatCompletionRequester) probeUsingAppropriateStandard(
	baseUrl string,
	llmConnection *models.LLMConnection,
	platformSettings settings.StandardSettingsInterface,
	apiKey *string,
) error {
	switch platformSettings.PlatformAPI() {
	case settings.PlatformAPIOpenAIResponsesStandard:
		if apiKey == nil {
			return fmt.Errorf("api key is required for OpenAI Responses Standard")
		}
		return r.openAIResponsesStandardChatCompletionRequester.Probe(baseUrl, *apiKey, llmConnection)

	case settings.PlatformAPIOpenAICompletionsStandard:
		if apiKey == nil {
			return fmt.Errorf("api key is required for OpenAI Completions Standard")
		}
		return r.openAICompletionsStandardChatCompletionRequester.Probe(baseUrl, *apiKey, llmConnection)

	case settings.PlatformAPIOllamaChatMessageStandard:
		return r.ollamaChatMessageStandardChatCompletionRequester.Probe(baseUrl, llmConnection)

	case settings.PlatformAPIAnthropicMessagesStandard:
		if apiKey == nil {
			return fmt.Errorf("api key is required for Anthropic Messages Standard")
		}
		return r.anthropicMessagesStandardChatCompletionRequester.Probe(baseUrl, *apiKey, llmConnection)

	case settings.PlatformAPIAmazonBedrockConverseStandard:
		if apiKey == nil {
			return fmt.Errorf("api key is required for Amazon Bedrock Converse Standard")
		}
		return r.amazonBedrockConverseStandardChatCompletionRequester.Probe(baseUrl, *apiKey, llmConnection)

	default:
		return fmt.Errorf("unsupported platform API: %s", platformSettings.PlatformAPI())
	}
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

// trimTokenizer resolves the gpt-4 tokenizer once; approximate for non-OpenAI models but cheap and consistent enough.
func (r *ChatCompletionRequester) trimTokenizer() (*tiktoken.Tiktoken, error) {
	r.tokenizerOnce.Do(func() {
		r.tokenizer, r.tokenizerErr = tiktoken.EncodingForModel("gpt-4")
	})
	return r.tokenizer, r.tokenizerErr
}

// trimMessagesToTokenLimit drops the oldest non-system messages until under
// limit; system messages stay at the head. Counts tokens with gpt-4's
// tokenizer for all platforms — close enough for trim.
func (r *ChatCompletionRequester) trimMessagesToTokenLimit(
	messages []*interfaces.ChatCompletionMessage,
	limit uint,
) ([]*interfaces.ChatCompletionMessage, error) {
	tokenizer, err := r.trimTokenizer()
	if err != nil {
		return nil, fmt.Errorf("failed to get tokenizer: %w", err)
	}

	systemMessages := make([]*interfaces.ChatCompletionMessage, 0, 1)
	rest := make([]*interfaces.ChatCompletionMessage, 0, len(messages))
	for _, msg := range messages {
		if msg.Role == "system" {
			systemMessages = append(systemMessages, msg)
		} else {
			rest = append(rest, msg)
		}
	}

	totalTokens := uint(0)
	for _, msg := range systemMessages {
		totalTokens += uint(len(tokenizer.Encode(msg.Content, nil, nil)))
	}
	restTokens := make([]uint, len(rest))
	for i, msg := range rest {
		restTokens[i] = uint(len(tokenizer.Encode(msg.Content, nil, nil)))
		totalTokens += restTokens[i]
	}

	startIdx := 0
	for totalTokens > limit && startIdx < len(rest) {
		totalTokens -= restTokens[startIdx]
		startIdx++
	}

	return append(systemMessages, rest[startIdx:]...), nil
}

func (r *ChatCompletionRequester) getPlatformSettings(
	platform enums.LLMConnectionPlatform,
) (settings.StandardSettingsInterface, error) {
	switch platform {
	case enums.LLMConnectionPlatformLibertAi:
		return r.libertAiSettings, nil
	case enums.LLMConnectionPlatformOpenAi:
		return r.openAiSettings, nil
	case enums.LLMConnectionPlatformOllama:
		return r.ollamaSettings, nil
	case enums.LLMConnectionPlatformMistralAi:
		return r.mistralSettings, nil
	case enums.LLMConnectionPlatformDeepseek:
		return r.deepSeekSettings, nil
	case enums.LLMConnectionPlatformXAi:
		return r.xSettings, nil
	case enums.LLMConnectionPlatformAnthropic:
		return r.anthropicSettings, nil
	case enums.LLMConnectionPlatformAwsBedrock:
		return r.amazonBedrockSettings, nil
	default:
		return nil, fmt.Errorf("unsupported platform: %s", platform)
	}
}

func (r *ChatCompletionRequester) completeUsingAppropriateStandard(
	baseUrl string,
	llmConnection *models.LLMConnection,
	selfTools []*models.Tool,
	mcpTools map[*models.MCPConnection][]*models.Tool,
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
			baseUrl,
			*apiKey,
			platformSettings.HasDirectAccessToMcp(),
			llmConnection,
			selfTools,
			mcpTools,
			messages,
		)

	case settings.PlatformAPIOpenAICompletionsStandard:
		if apiKey == nil {
			return nil, fmt.Errorf("api key is required for OpenAI Completions Standard")
		}
		return r.openAICompletionsStandardChatCompletionRequester.ChatCompletion(
			baseUrl,
			*apiKey,
			llmConnection,
			selfTools,
			mcpTools,
			messages,
		)

	case settings.PlatformAPIOllamaChatMessageStandard:
		return r.ollamaChatMessageStandardChatCompletionRequester.ChatCompletion(
			baseUrl,
			llmConnection,
			selfTools,
			mcpTools,
			messages,
		)

	case settings.PlatformAPIAnthropicMessagesStandard:
		if apiKey == nil {
			return nil, fmt.Errorf("api key is required for Anthropic Messages Standard")
		}
		return r.anthropicMessagesStandardChatCompletionRequester.ChatCompletion(
			baseUrl,
			*apiKey,
			platformSettings.HasDirectAccessToMcp(),
			llmConnection,
			selfTools,
			mcpTools,
			messages,
		)

	case settings.PlatformAPIAmazonBedrockConverseStandard:
		if apiKey == nil {
			return nil, fmt.Errorf("api key is required for Amazon Bedrock Converse Standard")
		}
		return r.amazonBedrockConverseStandardChatCompletionRequester.ChatCompletion(
			baseUrl,
			*apiKey,
			llmConnection,
			selfTools,
			mcpTools,
			messages,
		)

	default:
		return nil, fmt.Errorf("unsupported platform API: %s", platformSettings.PlatformAPI())
	}
}
