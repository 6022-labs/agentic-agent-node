package configurations

import (
	"net/http"

	"github.com/6022-labs/agentic-agent-node/src/agent_reasoning/services/interfaces"
	"github.com/6022-labs/agentic-agent-node/src/agent_reasoning_http_inference/clients"
	"github.com/6022-labs/agentic-agent-node/src/agent_reasoning_http_inference/metrics"
	metrics_interfaces "github.com/6022-labs/agentic-agent-node/src/agent_reasoning_http_inference/metrics/interfaces"
	http_inference_services "github.com/6022-labs/agentic-agent-node/src/agent_reasoning_http_inference/services"
	"github.com/6022-labs/agentic-agent-node/src/agent_reasoning_http_inference/settings"
	common_metrics "github.com/6022-labs/agentic-agent-node/src/common/metrics"
	"go.uber.org/dig"
)

func AddAgentReasoningHttpInferenceConfiguration(container *dig.Container) {
	// Settings
	err := container.Provide(settings.NewLibertAISettings)
	if err != nil {
		panic(err)
	}

	err = container.Provide(settings.NewOpenAISettings)
	if err != nil {
		panic(err)
	}

	err = container.Provide(settings.NewOllamaSettings)
	if err != nil {
		panic(err)
	}

	err = container.Provide(settings.NewXSettings)
	if err != nil {
		panic(err)
	}

	err = container.Provide(settings.NewDeepSeekSettings)
	if err != nil {
		panic(err)
	}

	err = container.Provide(settings.NewMistralSettings)
	if err != nil {
		panic(err)
	}

	err = container.Provide(settings.NewAnthropicSettings)
	if err != nil {
		panic(err)
	}

	err = container.Provide(settings.NewAmazonBedrockSettings)
	if err != nil {
		panic(err)
	}

	err = container.Provide(settings.NewTokenSettings)
	if err != nil {
		panic(err)
	}

	// Metrics
	err = container.Provide(
		metrics.NewLlmInferenceMetrics,
		dig.As(new(metrics_interfaces.LlmInferenceMetricsInterface)),
	)
	if err != nil {
		panic(err)
	}

	// Clients
	err = container.Provide(
		clients.NewOpenAICompletionsStandardClient,
		dig.As(new(clients.OpenAICompletionsStandardClientInterface)),
	)
	if err != nil {
		panic(err)
	}

	err = container.Provide(
		clients.NewOpenAIResponsesStandardClient,
		dig.As(new(clients.OpenAIResponsesStandardClientInterface)),
	)
	if err != nil {
		panic(err)
	}

	err = container.Provide(
		clients.NewOllamaChatMessageStandardClient,
		dig.As(new(clients.OllamaChatMessageStandardClientInterface)),
	)
	if err != nil {
		panic(err)
	}

	err = container.Provide(
		clients.NewAnthropicMessagesStandardClient,
		dig.As(new(clients.AnthropicMessagesStandardClientInterface)),
	)
	if err != nil {
		panic(err)
	}

	err = container.Provide(
		clients.NewAmazonBedrockConverseStandardClient,
		dig.As(new(clients.AmazonBedrockConverseStandardClientInterface)),
	)
	if err != nil {
		panic(err)
	}

	// Services
	err = container.Provide(
		http_inference_services.NewChatCompletionRequester,
		dig.As(new(interfaces.ChatCompletionRequesterInterface)),
	)
	if err != nil {
		panic(err)
	}

	err = container.Provide(
		http_inference_services.NewOpenAIResponsesStandardChatCompletionRequester,
		dig.As(new(http_inference_services.OpenAIResponsesStandardChatCompletionRequesterInterface)),
	)
	if err != nil {
		panic(err)
	}

	err = container.Provide(
		http_inference_services.NewOpenAICompletionsStandardChatCompletionRequester,
		dig.As(new(http_inference_services.OpenAICompletionsStandardChatCompletionRequesterInterface)),
	)
	if err != nil {
		panic(err)
	}

	err = container.Provide(
		http_inference_services.NewOllamaChatMessageStandardChatCompletionRequester,
		dig.As(new(http_inference_services.OllamaChatMessageStandardChatCompletionRequesterInterface)),
	)
	if err != nil {
		panic(err)
	}

	err = container.Provide(
		http_inference_services.NewAnthropicMessagesStandardChatCompletionRequester,
		dig.As(new(http_inference_services.AnthropicMessagesStandardChatCompletionRequesterInterface)),
	)
	if err != nil {
		panic(err)
	}

	err = container.Provide(
		http_inference_services.NewAmazonBedrockConverseStandardChatCompletionRequester,
		dig.As(new(http_inference_services.AmazonBedrockConverseStandardChatCompletionRequesterInterface)),
	)
	if err != nil {
		panic(err)
	}

	// Http clients — transport wrapped so raw calls also emit http.client.* metrics.
	err = container.Provide(func() *http.Client {
		return &http.Client{
			Transport: common_metrics.NewHttpMetricsRoundTripper(
				http.DefaultTransport,
				common_metrics.NewExternalHttpMetrics("agentic_runtime_http_inference"),
			)}
	}, dig.Name("llm-inference-http-client"))
	if err != nil {
		panic(err)
	}
}
