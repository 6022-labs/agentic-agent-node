package services_test

import (
	"context"
	"errors"
	"testing"

	"github.com/6022-labs/agentic-agent-node/src/agent_reasoning/models"
	"github.com/6022-labs/agentic-agent-node/src/agent_reasoning/models/enums"
	"github.com/6022-labs/agentic-agent-node/src/agent_reasoning_http_inference/services"
	"github.com/6022-labs/agentic-agent-node/src/agent_reasoning_http_inference/settings"
	metrics_interfaces_mocks "github.com/6022-labs/agentic-agent-node/tests/agent_reasoning_http_inference_mocks/metrics_mocks/interfaces_mocks"
	"github.com/6022-labs/agentic-agent-node/tests/agent_reasoning_http_inference_mocks/services_mocks"
	reasoning_interfaces_mocks "github.com/6022-labs/agentic-agent-node/tests/agent_reasoning_mocks/services_mocks/interfaces_mocks"
	"github.com/6022-labs/agentic-agent-node/tests/common_tests/model_builders"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
	"go.uber.org/zap"
)

type WhenProbingLLMConnectionTestingSuite struct {
	sut *services.ChatCompletionRequester

	xSettings         *settings.XSettings
	openAiSettings    *settings.OpenAISettings
	mistralSettings   *settings.MistralSettings
	deepSeekSettings  *settings.DeepSeekSettings
	libertAiSettings  *settings.LibertAISettings
	anthropicSettings *settings.AnthropicSettings

	mockLlmSecretStorage                                     *reasoning_interfaces_mocks.MockLLMSecretStorageInterface
	mockOpenAIResponsesStandardChatCompletionRequester       *services_mocks.MockOpenAIResponsesStandardChatCompletionRequesterInterface
	mockOllamaChatMessageStandardChatCompletionRequester     *services_mocks.MockOllamaChatMessageStandardChatCompletionRequesterInterface
	mockAnthropicMessagesStandardChatCompletionRequester     *services_mocks.MockAnthropicMessagesStandardChatCompletionRequesterInterface
	mockOpenAICompletionsStandardChatCompletionRequester     *services_mocks.MockOpenAICompletionsStandardChatCompletionRequesterInterface
	mockAmazonBedrockConverseStandardChatCompletionRequester *services_mocks.MockAmazonBedrockConverseStandardChatCompletionRequesterInterface
}

func WhenProbingLLMConnectionBeforeEach(t *testing.T) *WhenProbingLLMConnectionTestingSuite {
	mockController := gomock.NewController(t)

	mockLlmSecretStorage := reasoning_interfaces_mocks.NewMockLLMSecretStorageInterface(mockController)
	mockOpenAIResponsesStandardChatCompletionRequester := services_mocks.NewMockOpenAIResponsesStandardChatCompletionRequesterInterface(
		mockController,
	)
	mockOllamaChatMessageStandardChatCompletionRequester := services_mocks.NewMockOllamaChatMessageStandardChatCompletionRequesterInterface(
		mockController,
	)
	mockAnthropicMessagesStandardChatCompletionRequester := services_mocks.NewMockAnthropicMessagesStandardChatCompletionRequesterInterface(
		mockController,
	)
	mockOpenAICompletionsStandardChatCompletionRequester := services_mocks.NewMockOpenAICompletionsStandardChatCompletionRequesterInterface(
		mockController,
	)
	mockAmazonBedrockConverseStandardChatCompletionRequester := services_mocks.NewMockAmazonBedrockConverseStandardChatCompletionRequesterInterface(
		mockController,
	)

	mockMetrics := metrics_interfaces_mocks.NewMockLlmInferenceMetricsInterface(mockController)
	mockMetrics.EXPECT().RecordProbe(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).AnyTimes()

	xSettings := &settings.XSettings{}
	openAiSettings := &settings.OpenAISettings{}
	mistralSettings := &settings.MistralSettings{}
	deepSeekSettings := &settings.DeepSeekSettings{}
	libertAiSettings := &settings.LibertAISettings{}
	anthropicSettings := &settings.AnthropicSettings{}

	sut := services.NewChatCompletionRequester(
		zap.NewNop(),
		xSettings,
		&settings.TokenSettings{MaxTokensPerRequest: 4096},
		mockLlmSecretStorage,
		mockMetrics,
		openAiSettings,
		&settings.OllamaSettings{},
		mistralSettings,
		deepSeekSettings,
		libertAiSettings,
		anthropicSettings,
		&settings.AmazonBedrockSettings{},
		mockOpenAIResponsesStandardChatCompletionRequester,
		mockOllamaChatMessageStandardChatCompletionRequester,
		mockAnthropicMessagesStandardChatCompletionRequester,
		mockOpenAICompletionsStandardChatCompletionRequester,
		mockAmazonBedrockConverseStandardChatCompletionRequester,
	)

	return &WhenProbingLLMConnectionTestingSuite{
		sut: sut,

		xSettings:         xSettings,
		openAiSettings:    openAiSettings,
		mistralSettings:   mistralSettings,
		deepSeekSettings:  deepSeekSettings,
		libertAiSettings:  libertAiSettings,
		anthropicSettings: anthropicSettings,

		mockLlmSecretStorage: mockLlmSecretStorage,
		mockOpenAIResponsesStandardChatCompletionRequester:       mockOpenAIResponsesStandardChatCompletionRequester,
		mockOllamaChatMessageStandardChatCompletionRequester:     mockOllamaChatMessageStandardChatCompletionRequester,
		mockAnthropicMessagesStandardChatCompletionRequester:     mockAnthropicMessagesStandardChatCompletionRequester,
		mockOpenAICompletionsStandardChatCompletionRequester:     mockOpenAICompletionsStandardChatCompletionRequester,
		mockAmazonBedrockConverseStandardChatCompletionRequester: mockAmazonBedrockConverseStandardChatCompletionRequester,
	}
}

func TestWhenProbingLLMConnection(t *testing.T) {
	t.Parallel()

	apiKey := "api-key"

	t.Run("Given the LLM connection is not supported", func(t *testing.T) {
		t.Parallel()

		t.Run("Should return an error", func(t *testing.T) {
			t.Parallel()

			suite := WhenProbingLLMConnectionBeforeEach(t)

			llmConnection := model_builders.NewLLMConnectionBuilder().
				WithPlatform("unsupported").
				WithModel("claude").
				Build()

			err := suite.sut.ProbeConnection(context.Background(), llmConnection, apiKey)
			assert.ErrorContains(t, err, "unsupported platform")
		})
	})

	t.Run("Given no api key is provided and none is stored", func(t *testing.T) {
		t.Parallel()

		initSuite := func(suite *WhenProbingLLMConnectionTestingSuite, _ *models.LLMConnection) {
			suite.mockLlmSecretStorage.EXPECT().Get(gomock.Any()).Return("", false, nil)
		}

		t.Run("Should return an error without calling the provider", func(t *testing.T) {
			t.Parallel()

			suite := WhenProbingLLMConnectionBeforeEach(t)

			llmConnection := model_builders.NewLLMConnectionBuilder().
				WithPlatform(enums.LLMConnectionPlatformOpenAi).
				WithModel("gpt-4").
				Build()
			initSuite(suite, llmConnection)

			err := suite.sut.ProbeConnection(context.Background(), llmConnection, "")
			assert.ErrorContains(t, err, "api key is required")
		})
	})

	t.Run("Given the provider rejects the probe", func(t *testing.T) {
		t.Parallel()

		initSuite := func(suite *WhenProbingLLMConnectionTestingSuite, llmConnection *models.LLMConnection) {
			suite.mockOpenAIResponsesStandardChatCompletionRequester.
				EXPECT().
				Probe(gomock.Any(), apiKey, llmConnection).
				Return(errors.New("401 unauthorized"))
		}

		t.Run("Should return the provider error", func(t *testing.T) {
			t.Parallel()

			suite := WhenProbingLLMConnectionBeforeEach(t)

			llmConnection := model_builders.NewLLMConnectionBuilder().
				WithPlatform(enums.LLMConnectionPlatformOpenAi).
				WithModel("gpt-4").
				Build()
			initSuite(suite, llmConnection)

			err := suite.sut.ProbeConnection(context.Background(), llmConnection, apiKey)
			assert.ErrorContains(t, err, "401 unauthorized")
		})
	})

	t.Run("Given an api key is provided", func(t *testing.T) {
		t.Parallel()

		initSuite := func(suite *WhenProbingLLMConnectionTestingSuite, llmConnection *models.LLMConnection) {
			suite.openAiSettings.BaseUrl = "http://openai.api"

			suite.mockOpenAIResponsesStandardChatCompletionRequester.
				EXPECT().
				Probe(suite.openAiSettings.GetBaseUrl(), "provided-key", llmConnection).
				Return(nil)
		}

		t.Run("Should probe with that key without reading vault", func(t *testing.T) {
			t.Parallel()

			suite := WhenProbingLLMConnectionBeforeEach(t)

			llmConnection := model_builders.NewLLMConnectionBuilder().
				WithPlatform(enums.LLMConnectionPlatformOpenAi).
				WithModel("gpt-4").
				Build()
			initSuite(suite, llmConnection)

			err := suite.sut.ProbeConnection(context.Background(), llmConnection, "provided-key")
			assert.NoError(t, err)
		})
	})

	t.Run("Given the LLM connection is LibertAI", func(t *testing.T) {
		t.Parallel()

		initSuite := func(suite *WhenProbingLLMConnectionTestingSuite, llmConnection *models.LLMConnection) {
			suite.libertAiSettings.BaseUrl = "http://libertai.api"
			suite.mockLlmSecretStorage.EXPECT().Get(gomock.Any()).Return(apiKey, true, nil)

			suite.mockOpenAICompletionsStandardChatCompletionRequester.
				EXPECT().
				Probe(suite.libertAiSettings.GetBaseUrl(), apiKey, llmConnection).
				Return(nil)
		}

		t.Run("Should probe using OpenAI completions standard", func(t *testing.T) {
			t.Parallel()

			suite := WhenProbingLLMConnectionBeforeEach(t)

			llmConnection := model_builders.NewLLMConnectionBuilder().
				WithPlatform(enums.LLMConnectionPlatformLibertAi).
				WithModel("gpt-4").
				Build()
			initSuite(suite, llmConnection)

			err := suite.sut.ProbeConnection(context.Background(), llmConnection, "")
			assert.NoError(t, err)
		})
	})

	t.Run("Given the LLM connection is OpenAI", func(t *testing.T) {
		t.Parallel()

		initSuite := func(suite *WhenProbingLLMConnectionTestingSuite, llmConnection *models.LLMConnection) {
			suite.openAiSettings.BaseUrl = "http://openai.api"
			suite.mockLlmSecretStorage.EXPECT().Get(gomock.Any()).Return(apiKey, true, nil)

			suite.mockOpenAIResponsesStandardChatCompletionRequester.
				EXPECT().
				Probe(suite.openAiSettings.GetBaseUrl(), apiKey, llmConnection).
				Return(nil)
		}

		t.Run("Should probe using OpenAI responses standard", func(t *testing.T) {
			t.Parallel()

			suite := WhenProbingLLMConnectionBeforeEach(t)

			llmConnection := model_builders.NewLLMConnectionBuilder().
				WithPlatform(enums.LLMConnectionPlatformOpenAi).
				WithModel("gpt-4").
				Build()
			initSuite(suite, llmConnection)

			err := suite.sut.ProbeConnection(context.Background(), llmConnection, "")
			assert.NoError(t, err)
		})
	})

	t.Run("Given the LLM connection is Ollama", func(t *testing.T) {
		t.Parallel()

		initSuite := func(suite *WhenProbingLLMConnectionTestingSuite, llmConnection *models.LLMConnection) {
			suite.mockLlmSecretStorage.EXPECT().Get(gomock.Any()).Return("", false, nil)

			suite.mockOllamaChatMessageStandardChatCompletionRequester.
				EXPECT().
				Probe(*llmConnection.BaseUrl, llmConnection).
				Return(nil)
		}

		t.Run("Should probe using Ollama chat message standard", func(t *testing.T) {
			t.Parallel()

			suite := WhenProbingLLMConnectionBeforeEach(t)

			llmConnection := model_builders.NewLLMConnectionBuilder().
				WithPlatform(enums.LLMConnectionPlatformOllama).
				WithBaseUrl("http://ollama.api").
				WithModel("ollama-model").
				Build()
			initSuite(suite, llmConnection)

			err := suite.sut.ProbeConnection(context.Background(), llmConnection, "")
			assert.NoError(t, err)
		})
	})

	t.Run("Given the LLM connection is Mistral", func(t *testing.T) {
		t.Parallel()

		initSuite := func(suite *WhenProbingLLMConnectionTestingSuite, llmConnection *models.LLMConnection) {
			suite.mistralSettings.BaseUrl = "http://mistral.api"
			suite.mockLlmSecretStorage.EXPECT().Get(gomock.Any()).Return(apiKey, true, nil)

			suite.mockOpenAICompletionsStandardChatCompletionRequester.
				EXPECT().
				Probe(suite.mistralSettings.GetBaseUrl(), apiKey, llmConnection).
				Return(nil)
		}

		t.Run("Should probe using OpenAI completions standard", func(t *testing.T) {
			t.Parallel()

			suite := WhenProbingLLMConnectionBeforeEach(t)

			llmConnection := model_builders.NewLLMConnectionBuilder().
				WithPlatform(enums.LLMConnectionPlatformMistralAi).
				WithModel("mistral-large").
				Build()
			initSuite(suite, llmConnection)

			err := suite.sut.ProbeConnection(context.Background(), llmConnection, "")
			assert.NoError(t, err)
		})
	})

	t.Run("Given the LLM connection is DeepSeek", func(t *testing.T) {
		t.Parallel()

		initSuite := func(suite *WhenProbingLLMConnectionTestingSuite, llmConnection *models.LLMConnection) {
			suite.deepSeekSettings.BaseUrl = "http://deepseek.api"
			suite.mockLlmSecretStorage.EXPECT().Get(gomock.Any()).Return(apiKey, true, nil)

			suite.mockOpenAICompletionsStandardChatCompletionRequester.
				EXPECT().
				Probe(suite.deepSeekSettings.GetBaseUrl(), apiKey, llmConnection).
				Return(nil)
		}

		t.Run("Should probe using OpenAI completions standard", func(t *testing.T) {
			t.Parallel()

			suite := WhenProbingLLMConnectionBeforeEach(t)

			llmConnection := model_builders.NewLLMConnectionBuilder().
				WithPlatform(enums.LLMConnectionPlatformDeepseek).
				WithModel("deepseek-chat").
				Build()
			initSuite(suite, llmConnection)

			err := suite.sut.ProbeConnection(context.Background(), llmConnection, "")
			assert.NoError(t, err)
		})
	})

	t.Run("Given the LLM connection is X", func(t *testing.T) {
		t.Parallel()

		initSuite := func(suite *WhenProbingLLMConnectionTestingSuite, llmConnection *models.LLMConnection) {
			suite.xSettings.BaseUrl = "http://x.api"
			suite.mockLlmSecretStorage.EXPECT().Get(gomock.Any()).Return(apiKey, true, nil)

			suite.mockOpenAICompletionsStandardChatCompletionRequester.
				EXPECT().
				Probe(suite.xSettings.GetBaseUrl(), apiKey, llmConnection).
				Return(nil)
		}

		t.Run("Should probe using OpenAI completions standard", func(t *testing.T) {
			t.Parallel()

			suite := WhenProbingLLMConnectionBeforeEach(t)

			llmConnection := model_builders.NewLLMConnectionBuilder().
				WithPlatform(enums.LLMConnectionPlatformXAi).
				WithModel("grok").
				Build()
			initSuite(suite, llmConnection)

			err := suite.sut.ProbeConnection(context.Background(), llmConnection, "")
			assert.NoError(t, err)
		})
	})

	t.Run("Given the LLM connection is Anthropic", func(t *testing.T) {
		t.Parallel()

		initSuite := func(suite *WhenProbingLLMConnectionTestingSuite, llmConnection *models.LLMConnection) {
			suite.anthropicSettings.BaseUrl = "http://anthropic.api"
			suite.mockLlmSecretStorage.EXPECT().Get(gomock.Any()).Return(apiKey, true, nil)

			suite.mockAnthropicMessagesStandardChatCompletionRequester.
				EXPECT().
				Probe(suite.anthropicSettings.GetBaseUrl(), apiKey, llmConnection).
				Return(nil)
		}

		t.Run("Should probe using Anthropic messages standard", func(t *testing.T) {
			t.Parallel()

			suite := WhenProbingLLMConnectionBeforeEach(t)

			llmConnection := model_builders.NewLLMConnectionBuilder().
				WithPlatform(enums.LLMConnectionPlatformAnthropic).
				WithModel("claude").
				Build()
			initSuite(suite, llmConnection)

			err := suite.sut.ProbeConnection(context.Background(), llmConnection, "")
			assert.NoError(t, err)
		})
	})

	t.Run("Given the LLM connection is Amazon Bedrock", func(t *testing.T) {
		t.Parallel()

		initSuite := func(suite *WhenProbingLLMConnectionTestingSuite, llmConnection *models.LLMConnection) {
			suite.mockLlmSecretStorage.EXPECT().Get(gomock.Any()).Return(apiKey, true, nil)

			suite.mockAmazonBedrockConverseStandardChatCompletionRequester.
				EXPECT().
				Probe(*llmConnection.BaseUrl, apiKey, llmConnection).
				Return(nil)
		}

		t.Run("Should probe using Amazon Bedrock converse standard", func(t *testing.T) {
			t.Parallel()

			suite := WhenProbingLLMConnectionBeforeEach(t)

			llmConnection := model_builders.NewLLMConnectionBuilder().
				WithPlatform(enums.LLMConnectionPlatformAwsBedrock).
				WithModel("anthropic.claude").
				Build()
			initSuite(suite, llmConnection)

			err := suite.sut.ProbeConnection(context.Background(), llmConnection, "")
			assert.NoError(t, err)
		})
	})
}
