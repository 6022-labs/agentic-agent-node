package services_test

import (
	"context"
	"testing"

	"github.com/6022-labs/agentic-agent-node/src/agent_reasoning/models"
	"github.com/6022-labs/agentic-agent-node/src/agent_reasoning/models/enums"
	"github.com/6022-labs/agentic-agent-node/src/agent_reasoning/services/interfaces"
	"github.com/6022-labs/agentic-agent-node/src/agent_reasoning_http_inference/services"
	"github.com/6022-labs/agentic-agent-node/src/agent_reasoning_http_inference/settings"
	metrics_interfaces_mocks "github.com/6022-labs/agentic-agent-node/tests/agent_reasoning_http_inference_mocks/metrics_mocks/interfaces_mocks"
	"github.com/6022-labs/agentic-agent-node/tests/agent_reasoning_http_inference_mocks/services_mocks"
	reasoning_interfaces_mocks "github.com/6022-labs/agentic-agent-node/tests/agent_reasoning_mocks/services_mocks/interfaces_mocks"
	"github.com/6022-labs/agentic-agent-node/tests/common_tests/model_builders"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
	"go.uber.org/zap"
)

type WhenRequestingChatCompletionTestingSuite struct {
	sut *services.ChatCompletionRequester

	tokenSettings *settings.TokenSettings

	xSettings             *settings.XSettings
	openAiSettings        *settings.OpenAISettings
	mistralSettings       *settings.MistralSettings
	deepSeekSettings      *settings.DeepSeekSettings
	libertAiSettings      *settings.LibertAISettings
	anthropicSettings     *settings.AnthropicSettings
	amazonBedrockSettings *settings.AmazonBedrockSettings

	mockLlmSecretStorage                                     *reasoning_interfaces_mocks.MockLLMSecretStorageInterface
	mockOpenAIResponsesStandardChatCompletionRequester       *services_mocks.MockOpenAIResponsesStandardChatCompletionRequesterInterface
	mockOllamaChatMessageStandardChatCompletionRequester     *services_mocks.MockOllamaChatMessageStandardChatCompletionRequesterInterface
	mockAnthropicMessagesStandardChatCompletionRequester     *services_mocks.MockAnthropicMessagesStandardChatCompletionRequesterInterface
	mockOpenAICompletionsStandardChatCompletionRequester     *services_mocks.MockOpenAICompletionsStandardChatCompletionRequesterInterface
	mockAmazonBedrockConverseStandardChatCompletionRequester *services_mocks.MockAmazonBedrockConverseStandardChatCompletionRequesterInterface
}

func WhenRequestingChatCompletionBeforeEach(t *testing.T) *WhenRequestingChatCompletionTestingSuite {
	mockController := gomock.NewController(t)

	mockLlmSecretStorage := reasoning_interfaces_mocks.NewMockLLMSecretStorageInterface(mockController)
	mockOpenAIResponsesStandardChatCompletionRequester := services_mocks.NewMockOpenAIResponsesStandardChatCompletionRequesterInterface(mockController)
	mockOllamaChatMessageStandardChatCompletionRequester := services_mocks.NewMockOllamaChatMessageStandardChatCompletionRequesterInterface(mockController)
	mockAnthropicMessagesStandardChatCompletionRequester := services_mocks.NewMockAnthropicMessagesStandardChatCompletionRequesterInterface(mockController)
	mockOpenAICompletionsStandardChatCompletionRequester := services_mocks.NewMockOpenAICompletionsStandardChatCompletionRequesterInterface(mockController)
	mockAmazonBedrockConverseStandardChatCompletionRequester := services_mocks.NewMockAmazonBedrockConverseStandardChatCompletionRequesterInterface(mockController)

	mockMetrics := metrics_interfaces_mocks.NewMockLlmInferenceMetricsInterface(mockController)
	mockMetrics.EXPECT().RecordOperationDuration(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).AnyTimes()
	mockMetrics.EXPECT().RecordTokenUsage(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).AnyTimes()

	tokenSettings := &settings.TokenSettings{MaxTokensPerRequest: 4096}
	xSettings := &settings.XSettings{}
	openAiSettings := &settings.OpenAISettings{}
	ollamaSettings := &settings.OllamaSettings{}
	mistralSettings := &settings.MistralSettings{}
	deepSeekSettings := &settings.DeepSeekSettings{}
	libertAiSettings := &settings.LibertAISettings{}
	anthropicSettings := &settings.AnthropicSettings{}
	amazonBedrockSettings := &settings.AmazonBedrockSettings{}

	sut := services.NewChatCompletionRequester(
		zap.NewNop(),
		xSettings,
		tokenSettings,
		mockLlmSecretStorage,
		mockMetrics,
		openAiSettings,
		ollamaSettings,
		mistralSettings,
		deepSeekSettings,
		libertAiSettings,
		anthropicSettings,
		amazonBedrockSettings,
		mockOpenAIResponsesStandardChatCompletionRequester,
		mockOllamaChatMessageStandardChatCompletionRequester,
		mockAnthropicMessagesStandardChatCompletionRequester,
		mockOpenAICompletionsStandardChatCompletionRequester,
		mockAmazonBedrockConverseStandardChatCompletionRequester,
	)

	return &WhenRequestingChatCompletionTestingSuite{
		sut: sut,

		tokenSettings: tokenSettings,

		xSettings:             xSettings,
		openAiSettings:        openAiSettings,
		mistralSettings:       mistralSettings,
		deepSeekSettings:      deepSeekSettings,
		libertAiSettings:      libertAiSettings,
		anthropicSettings:     anthropicSettings,
		amazonBedrockSettings: amazonBedrockSettings,

		mockLlmSecretStorage: mockLlmSecretStorage,
		mockOpenAIResponsesStandardChatCompletionRequester:       mockOpenAIResponsesStandardChatCompletionRequester,
		mockOllamaChatMessageStandardChatCompletionRequester:     mockOllamaChatMessageStandardChatCompletionRequester,
		mockAnthropicMessagesStandardChatCompletionRequester:     mockAnthropicMessagesStandardChatCompletionRequester,
		mockOpenAICompletionsStandardChatCompletionRequester:     mockOpenAICompletionsStandardChatCompletionRequester,
		mockAmazonBedrockConverseStandardChatCompletionRequester: mockAmazonBedrockConverseStandardChatCompletionRequester,
	}
}

func TestWhenRequestingChatCompletion(t *testing.T) {
	t.Parallel()

	apiKey := "api-key"
	completionText := "completion-result"
	completion := &interfaces.LLMCompletion{Text: &completionText}
	messages := []*interfaces.ChatCompletionMessage{{Role: "user", Content: "hello"}}

	t.Run("Given llm connection is LibertAI", func(t *testing.T) {
		t.Parallel()

		initSuite := func(suite *WhenRequestingChatCompletionTestingSuite, llmConnection *models.LLMConnection) {
			suite.libertAiSettings.BaseUrl = "http://libertai.api"
			suite.mockLlmSecretStorage.EXPECT().Get(gomock.Any()).Return(apiKey, true, nil)

			suite.mockOpenAICompletionsStandardChatCompletionRequester.
				EXPECT().
				ChatCompletion(
					suite.libertAiSettings.GetBaseUrl(),
					apiKey,
					llmConnection,
					gomock.Any(),
					gomock.Any(),
					gomock.Any(),
				).
				Return(completion, nil)
		}

		t.Run("Should use OpenAI completions standard", func(t *testing.T) {
			t.Parallel()

			suite := WhenRequestingChatCompletionBeforeEach(t)

			llmConnection := model_builders.NewLLMConnectionBuilder().
				WithPlatform(enums.LLMConnectionPlatformLibertAi).
				WithModel("gpt-4").
				Build()
			initSuite(suite, llmConnection)

			result, err := suite.sut.ChatCompletion(context.Background(), llmConnection, nil, nil, messages)
			assert.NoError(t, err)
			assert.Equal(t, completion, result)
		})
	})

	t.Run("Given llm connection is OpenAI", func(t *testing.T) {
		t.Parallel()

		initSuite := func(suite *WhenRequestingChatCompletionTestingSuite, llmConnection *models.LLMConnection) {
			suite.openAiSettings.BaseUrl = "http://openai.api"
			suite.mockLlmSecretStorage.EXPECT().Get(gomock.Any()).Return(apiKey, true, nil)

			suite.mockOpenAIResponsesStandardChatCompletionRequester.
				EXPECT().
				ChatCompletion(
					suite.openAiSettings.GetBaseUrl(),
					apiKey,
					suite.openAiSettings.HasDirectAccessToMcp(),
					llmConnection,
					gomock.Any(),
					gomock.Any(),
					gomock.Any(),
				).
				Return(completion, nil)
		}

		t.Run("Should use OpenAI responses standard", func(t *testing.T) {
			t.Parallel()

			suite := WhenRequestingChatCompletionBeforeEach(t)

			llmConnection := model_builders.NewLLMConnectionBuilder().
				WithPlatform(enums.LLMConnectionPlatformOpenAi).
				WithModel("gpt-4").
				Build()
			initSuite(suite, llmConnection)

			result, err := suite.sut.ChatCompletion(context.Background(), llmConnection, nil, nil, messages)
			assert.NoError(t, err)
			assert.Equal(t, completion, result)
		})
	})

	t.Run("Given llm connection is Ollama", func(t *testing.T) {
		t.Parallel()

		initSuite := func(suite *WhenRequestingChatCompletionTestingSuite, llmConnection *models.LLMConnection) {
			suite.mockLlmSecretStorage.EXPECT().Get(gomock.Any()).Return("", false, nil)

			suite.mockOllamaChatMessageStandardChatCompletionRequester.
				EXPECT().
				ChatCompletion(
					*llmConnection.BaseUrl,
					llmConnection,
					gomock.Any(),
					gomock.Any(),
					gomock.Any(),
				).
				Return(completion, nil)
		}

		t.Run("Should use Ollama chat message standard", func(t *testing.T) {
			t.Parallel()

			suite := WhenRequestingChatCompletionBeforeEach(t)

			llmConnection := model_builders.NewLLMConnectionBuilder().
				WithPlatform(enums.LLMConnectionPlatformOllama).
				WithBaseUrl("http://ollama.api").
				WithModel("ollama-model").
				Build()
			initSuite(suite, llmConnection)

			result, err := suite.sut.ChatCompletion(context.Background(), llmConnection, nil, nil, messages)
			assert.NoError(t, err)
			assert.Equal(t, completion, result)
		})
	})

	t.Run("Given llm connection is Mistral", func(t *testing.T) {
		t.Parallel()

		initSuite := func(suite *WhenRequestingChatCompletionTestingSuite, llmConnection *models.LLMConnection) {
			suite.mistralSettings.BaseUrl = "http://mistral.api"
			suite.mockLlmSecretStorage.EXPECT().Get(gomock.Any()).Return(apiKey, true, nil)

			suite.mockOpenAICompletionsStandardChatCompletionRequester.
				EXPECT().
				ChatCompletion(
					suite.mistralSettings.GetBaseUrl(),
					apiKey,
					llmConnection,
					gomock.Any(),
					gomock.Any(),
					gomock.Any(),
				).
				Return(completion, nil)
		}

		t.Run("Should use OpenAI completions standard", func(t *testing.T) {
			t.Parallel()

			suite := WhenRequestingChatCompletionBeforeEach(t)

			llmConnection := model_builders.NewLLMConnectionBuilder().
				WithPlatform(enums.LLMConnectionPlatformMistralAi).
				WithModel("mistral-7b").
				Build()
			initSuite(suite, llmConnection)

			result, err := suite.sut.ChatCompletion(context.Background(), llmConnection, nil, nil, messages)
			assert.NoError(t, err)
			assert.Equal(t, completion, result)
		})
	})

	t.Run("Given llm connection is DeepSeek", func(t *testing.T) {
		t.Parallel()

		initSuite := func(suite *WhenRequestingChatCompletionTestingSuite, llmConnection *models.LLMConnection) {
			suite.deepSeekSettings.BaseUrl = "http://deepseek.api"
			suite.mockLlmSecretStorage.EXPECT().Get(gomock.Any()).Return(apiKey, true, nil)

			suite.mockOpenAICompletionsStandardChatCompletionRequester.
				EXPECT().
				ChatCompletion(
					suite.deepSeekSettings.GetBaseUrl(),
					apiKey,
					llmConnection,
					gomock.Any(),
					gomock.Any(),
					gomock.Any(),
				).
				Return(completion, nil)
		}

		t.Run("Should use OpenAI completions standard", func(t *testing.T) {
			t.Parallel()

			suite := WhenRequestingChatCompletionBeforeEach(t)

			llmConnection := model_builders.NewLLMConnectionBuilder().
				WithPlatform(enums.LLMConnectionPlatformDeepseek).
				WithModel("deepseek-7b").
				Build()
			initSuite(suite, llmConnection)

			result, err := suite.sut.ChatCompletion(context.Background(), llmConnection, nil, nil, messages)
			assert.NoError(t, err)
			assert.Equal(t, completion, result)
		})
	})

	t.Run("Given llm connection is X", func(t *testing.T) {
		t.Parallel()

		initSuite := func(suite *WhenRequestingChatCompletionTestingSuite, llmConnection *models.LLMConnection) {
			suite.xSettings.BaseUrl = "http://x.api"
			suite.mockLlmSecretStorage.EXPECT().Get(gomock.Any()).Return(apiKey, true, nil)

			suite.mockOpenAICompletionsStandardChatCompletionRequester.
				EXPECT().
				ChatCompletion(
					suite.xSettings.GetBaseUrl(),
					apiKey,
					llmConnection,
					gomock.Any(),
					gomock.Any(),
					gomock.Any(),
				).
				Return(completion, nil)
		}

		t.Run("Should use OpenAI completions standard", func(t *testing.T) {
			t.Parallel()

			suite := WhenRequestingChatCompletionBeforeEach(t)

			llmConnection := model_builders.NewLLMConnectionBuilder().
				WithPlatform(enums.LLMConnectionPlatformXAi).
				WithModel("grok").
				Build()
			initSuite(suite, llmConnection)

			result, err := suite.sut.ChatCompletion(context.Background(), llmConnection, nil, nil, messages)
			assert.NoError(t, err)
			assert.Equal(t, completion, result)
		})
	})

	t.Run("Given llm connection is Anthropic", func(t *testing.T) {
		t.Parallel()

		initSuite := func(suite *WhenRequestingChatCompletionTestingSuite, llmConnection *models.LLMConnection) {
			suite.anthropicSettings.BaseUrl = "http://anthropic.api"
			suite.mockLlmSecretStorage.EXPECT().Get(gomock.Any()).Return(apiKey, true, nil)

			suite.mockAnthropicMessagesStandardChatCompletionRequester.
				EXPECT().
				ChatCompletion(
					suite.anthropicSettings.GetBaseUrl(),
					apiKey,
					suite.anthropicSettings.HasDirectAccessToMcp(),
					llmConnection,
					gomock.Any(),
					gomock.Any(),
					gomock.Any(),
				).
				Return(completion, nil)
		}

		t.Run("Should use Anthropic messages standard", func(t *testing.T) {
			t.Parallel()

			suite := WhenRequestingChatCompletionBeforeEach(t)

			llmConnection := model_builders.NewLLMConnectionBuilder().
				WithPlatform(enums.LLMConnectionPlatformAnthropic).
				WithModel("claude").
				Build()
			initSuite(suite, llmConnection)

			result, err := suite.sut.ChatCompletion(context.Background(), llmConnection, nil, nil, messages)
			assert.NoError(t, err)
			assert.Equal(t, completion, result)
		})
	})

	t.Run("Given llm connection is Amazon Bedrock", func(t *testing.T) {
		t.Parallel()

		initSuite := func(suite *WhenRequestingChatCompletionTestingSuite, llmConnection *models.LLMConnection) {
			suite.mockLlmSecretStorage.EXPECT().Get(gomock.Any()).Return(apiKey, true, nil)

			suite.mockAmazonBedrockConverseStandardChatCompletionRequester.
				EXPECT().
				ChatCompletion(
					*llmConnection.BaseUrl,
					apiKey,
					llmConnection,
					gomock.Any(),
					gomock.Any(),
					gomock.Any(),
				).
				Return(completion, nil)
		}

		t.Run("Should use Amazon Bedrock converse standard", func(t *testing.T) {
			t.Parallel()

			suite := WhenRequestingChatCompletionBeforeEach(t)

			llmConnection := model_builders.NewLLMConnectionBuilder().
				WithPlatform(enums.LLMConnectionPlatformAwsBedrock).
				WithModel("anthropic.claude").
				Build()
			initSuite(suite, llmConnection)

			result, err := suite.sut.ChatCompletion(context.Background(), llmConnection, nil, nil, messages)
			assert.NoError(t, err)
			assert.Equal(t, completion, result)
		})
	})

	t.Run("Given the conversation exceeds the token limit", func(t *testing.T) {
		t.Parallel()

		overLimitMessages := []*interfaces.ChatCompletionMessage{
			{Role: "system", Content: "layered system prompt"},
			{Role: "user", Content: "oldest turn that should be dropped first"},
			{Role: "assistant", Content: "old answer"},
			{Role: "user", Content: "newest turn"},
		}

		initSuite := func(suite *WhenRequestingChatCompletionTestingSuite, llmConnection *models.LLMConnection, captured *[]*interfaces.ChatCompletionMessage) {
			suite.tokenSettings.MaxTokensPerRequest = 12
			suite.libertAiSettings.BaseUrl = "http://libertai.api"
			suite.mockLlmSecretStorage.EXPECT().Get(gomock.Any()).Return(apiKey, true, nil)

			suite.mockOpenAICompletionsStandardChatCompletionRequester.
				EXPECT().
				ChatCompletion(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
				DoAndReturn(func(_, _ string, _ *models.LLMConnection, _ []*models.MCPConnection, _ map[uuid.UUID][]*interfaces.McpTool, msgs []*interfaces.ChatCompletionMessage) (*interfaces.LLMCompletion, error) {
					*captured = msgs
					return completion, nil
				})
		}

		t.Run("Should trim oldest turns but always keep the system prompt", func(t *testing.T) {
			t.Parallel()

			suite := WhenRequestingChatCompletionBeforeEach(t)

			llmConnection := model_builders.NewLLMConnectionBuilder().
				WithPlatform(enums.LLMConnectionPlatformLibertAi).
				WithModel("gpt-4").
				Build()
			var captured []*interfaces.ChatCompletionMessage
			initSuite(suite, llmConnection, &captured)

			_, err := suite.sut.ChatCompletion(context.Background(), llmConnection, nil, nil, overLimitMessages)

			assert.NoError(t, err)
			if assert.NotEmpty(t, captured) {
				assert.Equal(t, "system", captured[0].Role)
				assert.Equal(t, "layered system prompt", captured[0].Content)
				for _, m := range captured {
					assert.NotEqual(t, "oldest turn that should be dropped first", m.Content)
				}
				assert.Equal(t, "newest turn", captured[len(captured)-1].Content)
			}
		})
	})

	t.Run("Given llm connection is not supported", func(t *testing.T) {
		t.Parallel()

		initSuite := func(suite *WhenRequestingChatCompletionTestingSuite, llmConnection *models.LLMConnection) {}

		t.Run("Should return an error", func(t *testing.T) {
			t.Parallel()

			suite := WhenRequestingChatCompletionBeforeEach(t)

			llmConnection := model_builders.NewLLMConnectionBuilder().
				WithPlatform("unsupported").
				WithModel("claude").
				Build()
			initSuite(suite, llmConnection)

			result, err := suite.sut.ChatCompletion(context.Background(), llmConnection, nil, nil, messages)
			assert.Error(t, err)
			assert.Nil(t, result)
		})
	})
}
