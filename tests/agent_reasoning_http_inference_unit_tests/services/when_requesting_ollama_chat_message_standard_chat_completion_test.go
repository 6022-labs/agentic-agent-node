package services_test

import (
	"encoding/json"
	"fmt"
	"testing"

	"github.com/6022-labs/agentic-agent-node/src/agent_reasoning/models"
	"github.com/6022-labs/agentic-agent-node/src/agent_reasoning/services/interfaces"
	external_models "github.com/6022-labs/agentic-agent-node/src/agent_reasoning_http_inference/models"
	"github.com/6022-labs/agentic-agent-node/src/agent_reasoning_http_inference/services"
	"github.com/6022-labs/agentic-agent-node/tests/agent_reasoning_http_inference_mocks/clients_mocks"
	"github.com/6022-labs/agentic-agent-node/tests/common_tests/model_builders"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
	"go.uber.org/zap"
)

type WhenRequestingOllamaChatMessageStandardChatCompletionTestingSuite struct {
	sut *services.OllamaChatMessageStandardChatCompletionRequester

	mockOllamaClient *clients_mocks.MockOllamaChatMessageStandardClientInterface
}

func WhenRequestingOllamaChatMessageStandardChatCompletionBeforeEach(
	t *testing.T,
) *WhenRequestingOllamaChatMessageStandardChatCompletionTestingSuite {
	mockController := gomock.NewController(t)

	mockOllamaClient := clients_mocks.NewMockOllamaChatMessageStandardClientInterface(mockController)

	sut := services.NewOllamaChatMessageStandardChatCompletionRequester(
		zap.NewNop(),
		mockOllamaClient,
	)

	return &WhenRequestingOllamaChatMessageStandardChatCompletionTestingSuite{
		sut:              sut,
		mockOllamaClient: mockOllamaClient,
	}
}

func TestWhenRequestingOllamaChatMessageStandardChatCompletion(t *testing.T) {
	t.Parallel()

	const baseUrl = "https://ollama.example.com"

	llmConnection := model_builders.NewLLMConnectionBuilder().
		WithModel("llama3").
		Build()

	messages := []*interfaces.ChatCompletionMessage{
		{Role: "user", Content: "hello"},
	}

	t.Run("Given the LLM returns plain text", func(t *testing.T) {
		t.Parallel()

		expectedContent := "hi there"

		initSuite := func(suite *WhenRequestingOllamaChatMessageStandardChatCompletionTestingSuite) {
			suite.mockOllamaClient.EXPECT().
				ChatMessage(baseUrl, gomock.Any()).
				DoAndReturn(func(_ string, request *external_models.ExternalOllamaChatMessageStandardRequest) (*external_models.ExternalOllamaChatMessageStandardResponse, error) {
					assert.Equal(t, llmConnection.Model, request.Model)
					if assert.Len(t, request.Messages, 1) {
						assert.Equal(t, "user", request.Messages[0].Role)
						assert.Equal(t, "hello", request.Messages[0].Content)
					}
					assert.Nil(t, request.Tools)

					return &external_models.ExternalOllamaChatMessageStandardResponse{
						Model: llmConnection.Model,
						Message: external_models.ExternalOllamaChatMessageStandardMessage{
							Role:    "assistant",
							Content: expectedContent,
						},
					}, nil
				})
		}

		t.Run("Should return the message content", func(t *testing.T) {
			t.Parallel()

			suite := WhenRequestingOllamaChatMessageStandardChatCompletionBeforeEach(t)
			initSuite(suite)

			result, err := suite.sut.ChatCompletion(baseUrl, llmConnection, nil, nil, messages, nil)
			assert.NoError(t, err)
			if assert.NotNil(t, result) {
				assert.Equal(t, []string{expectedContent}, result.Texts)
			}
		})
	})

	t.Run("Given a blank response content", func(t *testing.T) {
		t.Parallel()

		doneReason := "length"

		initSuite := func(suite *WhenRequestingOllamaChatMessageStandardChatCompletionTestingSuite) {
			suite.mockOllamaClient.EXPECT().
				ChatMessage(baseUrl, gomock.Any()).
				Return(&external_models.ExternalOllamaChatMessageStandardResponse{
					Model: llmConnection.Model,
					Message: external_models.ExternalOllamaChatMessageStandardMessage{
						Role:    "assistant",
						Content: "",
					},
					DoneReason: &doneReason,
				}, nil)
		}

		t.Run("Should return an error", func(t *testing.T) {
			t.Parallel()

			suite := WhenRequestingOllamaChatMessageStandardChatCompletionBeforeEach(t)
			initSuite(suite)

			result, err := suite.sut.ChatCompletion(baseUrl, llmConnection, nil, nil, messages, nil)
			assert.Nil(t, result)
			assert.ErrorContains(t, err, "ollama returned empty content")
			assert.ErrorContains(t, err, doneReason)
		})
	})

	t.Run("Given a self tool with no required fields", func(t *testing.T) {
		t.Parallel()

		selfTools := []*models.Tool{{Name: "get_identity", Description: "read own identity"}}
		expectedContent := "ok"

		initSuite := func(suite *WhenRequestingOllamaChatMessageStandardChatCompletionTestingSuite) {
			suite.mockOllamaClient.EXPECT().
				ChatMessage(baseUrl, gomock.Any()).
				DoAndReturn(func(_ string, request *external_models.ExternalOllamaChatMessageStandardRequest) (*external_models.ExternalOllamaChatMessageStandardResponse, error) {
					if assert.Len(t, request.Tools, 1) {
						assert.Equal(t, "self_get_identity", request.Tools[0].Function.Name)
						// Must serialize required as [], not the invalid "required":null.
						encoded, _ := json.Marshal(request.Tools[0].Function.Parameters)
						assert.Contains(t, string(encoded), `"required":[]`)
					}

					return &external_models.ExternalOllamaChatMessageStandardResponse{
						Model: llmConnection.Model,
						Message: external_models.ExternalOllamaChatMessageStandardMessage{
							Role:    "assistant",
							Content: expectedContent,
						},
					}, nil
				})
		}

		t.Run("Should serialize required as an empty array", func(t *testing.T) {
			t.Parallel()

			suite := WhenRequestingOllamaChatMessageStandardChatCompletionBeforeEach(t)
			initSuite(suite)

			result, err := suite.sut.ChatCompletion(baseUrl, llmConnection, selfTools, nil, messages, nil)
			assert.NoError(t, err)
			assert.NotNil(t, result)
		})
	})

	t.Run("Given the LLM emits a tool call", func(t *testing.T) {
		t.Parallel()

		mcpConnection := model_builders.NewMCPConnectionBuilder().
			WithName("inventory").
			WithUrl("https://mcp.example.com").
			Build()
		mcpTool := &models.Tool{Name: "fetch_items"}
		functionName := fmt.Sprintf("mcp_%s_%s", mcpConnection.Name, mcpTool.Name)
		toolArgs := map[string]any{"query": "value"}

		mcpTools := map[*models.MCPConnection][]*models.Tool{
			mcpConnection: {mcpTool},
		}

		initSuite := func(suite *WhenRequestingOllamaChatMessageStandardChatCompletionTestingSuite) {
			suite.mockOllamaClient.EXPECT().
				ChatMessage(baseUrl, gomock.Any()).
				Return(&external_models.ExternalOllamaChatMessageStandardResponse{
					Model: llmConnection.Model,
					Message: external_models.ExternalOllamaChatMessageStandardMessage{
						Role: "assistant",
						ToolCalls: []external_models.ExternalOllamaChatMessageStandardMessageToolCall{
							{
								Function: &external_models.ExternalOllamaChatMessageStandardMessageToolCallFunction{
									Name:      functionName,
									Arguments: toolArgs,
								},
							},
						},
					},
				}, nil)
		}

		t.Run("Should return the ToolCall (id == function name)", func(t *testing.T) {
			t.Parallel()

			suite := WhenRequestingOllamaChatMessageStandardChatCompletionBeforeEach(t)
			initSuite(suite)

			result, err := suite.sut.ChatCompletion(baseUrl, llmConnection, nil, mcpTools, messages, nil)
			assert.NoError(t, err)
			if assert.NotNil(t, result) && assert.Len(t, result.ToolCalls, 1) {
				assert.Equal(t, functionName, result.ToolCalls[0].Id)
				assert.Equal(t, functionName, result.ToolCalls[0].Name)
				assert.Equal(t, toolArgs, result.ToolCalls[0].Arguments)
			}
		})
	})
}
