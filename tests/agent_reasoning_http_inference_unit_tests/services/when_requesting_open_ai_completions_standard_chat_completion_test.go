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

type WhenRequestingOpenAICompletionsStandardChatCompletionTestingSuite struct {
	sut *services.OpenAICompletionsStandardChatCompletionRequester

	mockOpenAICompletionsClient *clients_mocks.MockOpenAICompletionsStandardClientInterface
}

func WhenRequestingOpenAICompletionsStandardChatCompletionBeforeEach(
	t *testing.T,
) *WhenRequestingOpenAICompletionsStandardChatCompletionTestingSuite {
	mockController := gomock.NewController(t)

	mockOpenAICompletionsClient := clients_mocks.NewMockOpenAICompletionsStandardClientInterface(mockController)

	sut := services.NewOpenAICompletionsStandardChatCompletionRequester(
		zap.NewNop(),
		mockOpenAICompletionsClient,
	)

	return &WhenRequestingOpenAICompletionsStandardChatCompletionTestingSuite{
		sut:                         sut,
		mockOpenAICompletionsClient: mockOpenAICompletionsClient,
	}
}

func TestWhenRequestingOpenAICompletionsStandardChatCompletion(t *testing.T) {
	t.Parallel()

	const (
		baseUrl = "https://openai.example.com"
		apiKey  = "test-key"

		expectedContent = "hi there"
	)

	llmConnection := model_builders.NewLLMConnectionBuilder().
		WithModel("gpt-4").
		Build()

	messages := []*interfaces.ChatCompletionMessage{
		{Role: "user", Content: "hello"},
	}

	t.Run("Given the response has no choices", func(t *testing.T) {
		t.Parallel()

		initSuite := func(suite *WhenRequestingOpenAICompletionsStandardChatCompletionTestingSuite) {
			suite.mockOpenAICompletionsClient.
				EXPECT().
				ChatCompletion(baseUrl, apiKey, gomock.Any()).
				Return(&external_models.ExternalOpenAICompletionsStandardResponse{
					ID:      "resp_1",
					Model:   llmConnection.Model,
					Choices: []external_models.ExternalOpenAICompletionsStandardResponseChoice{},
				}, nil)
		}

		t.Run("Should return an error", func(t *testing.T) {
			t.Parallel()

			suite := WhenRequestingOpenAICompletionsStandardChatCompletionBeforeEach(t)
			initSuite(suite)

			result, err := suite.sut.ChatCompletion(baseUrl, apiKey, llmConnection, nil, nil, messages, nil)
			assert.Nil(t, result)
			assert.ErrorContains(t, err, "openai completions returned no choices")
		})
	})

	t.Run("Given the input stream replays an earlier tool turn", func(t *testing.T) {
		t.Parallel()

		toolArguments := map[string]any{"query": "value"}
		argsJson, _ := json.Marshal(toolArguments)

		replay := []*interfaces.ChatCompletionMessage{
			{Role: "user", Content: "go"},
			{Role: "assistant", ToolCalls: []interfaces.ToolCall{
				{Id: "call_1", Name: "inventory_fetch_items", Arguments: toolArguments},
			}},
			{Role: "tool", ToolCallId: "call_1", Content: `{"status":"ok"}`},
		}

		initSuite := func(suite *WhenRequestingOpenAICompletionsStandardChatCompletionTestingSuite) {
			suite.mockOpenAICompletionsClient.
				EXPECT().
				ChatCompletion(baseUrl, apiKey, gomock.Any()).
				DoAndReturn(func(_ string, _ string, request *external_models.ExternalOpenAICompletionsStandardRequest) (*external_models.ExternalOpenAICompletionsStandardResponse, error) {
					if assert.Len(t, request.Messages, 3) {
						assistant := request.Messages[1]
						assert.Equal(t, "assistant", assistant.Role)
						if assert.Len(t, assistant.ToolCalls, 1) {
							assert.Equal(t, "call_1", assistant.ToolCalls[0].ID)
							assert.Equal(t, "inventory_fetch_items", assistant.ToolCalls[0].Function.Name)
							assert.Equal(t, string(argsJson), assistant.ToolCalls[0].Function.Arguments)
						}

						tool := request.Messages[2]
						assert.Equal(t, "tool", tool.Role)
						if assert.NotNil(t, tool.ToolCallId) {
							assert.Equal(t, "call_1", *tool.ToolCallId)
						}
					}

					return &external_models.ExternalOpenAICompletionsStandardResponse{
						Choices: []external_models.ExternalOpenAICompletionsStandardResponseChoice{
							{
								Message: external_models.ExternalOpenAICompletionsStandardResponseChoiceMessage{
									Role:    "assistant",
									Content: expectedContent,
								},
							},
						},
					}, nil
				})
		}

		t.Run("Should encode the assistant tool_calls + tool message in the request", func(t *testing.T) {
			t.Parallel()

			suite := WhenRequestingOpenAICompletionsStandardChatCompletionBeforeEach(t)
			initSuite(suite)

			result, err := suite.sut.ChatCompletion(baseUrl, apiKey, llmConnection, nil, nil, replay, nil)
			assert.NoError(t, err)
			if assert.NotNil(t, result) {
				assert.Equal(t, []string{expectedContent}, result.Texts)
			}
		})
	})

	t.Run("Given a self tool with no required fields", func(t *testing.T) {
		t.Parallel()

		selfTools := []*models.Tool{{Name: "get_identity", Description: "read own identity"}}

		initSuite := func(suite *WhenRequestingOpenAICompletionsStandardChatCompletionTestingSuite) {
			suite.mockOpenAICompletionsClient.
				EXPECT().
				ChatCompletion(baseUrl, apiKey, gomock.Any()).
				DoAndReturn(func(_ string, _ string, request *external_models.ExternalOpenAICompletionsStandardRequest) (*external_models.ExternalOpenAICompletionsStandardResponse, error) {
					if assert.Len(t, request.Tools, 1) {
						assert.Equal(t, "self_get_identity", request.Tools[0].Function.Name)
						// Must serialize required as [], not the invalid "required":null.
						encoded, _ := json.Marshal(request.Tools[0].Function.Parameters)
						assert.Contains(t, string(encoded), `"required":[]`)
					}

					return &external_models.ExternalOpenAICompletionsStandardResponse{
						Choices: []external_models.ExternalOpenAICompletionsStandardResponseChoice{
							{
								Message: external_models.ExternalOpenAICompletionsStandardResponseChoiceMessage{
									Role:    "assistant",
									Content: expectedContent,
								},
							},
						},
					}, nil
				})
		}

		t.Run("Should serialize required as an empty array", func(t *testing.T) {
			t.Parallel()

			suite := WhenRequestingOpenAICompletionsStandardChatCompletionBeforeEach(t)
			initSuite(suite)

			result, err := suite.sut.ChatCompletion(baseUrl, apiKey, llmConnection, selfTools, nil, messages, nil)
			assert.NoError(t, err)
			assert.NotNil(t, result)
		})
	})

	t.Run("Given the LLM returns plain text", func(t *testing.T) {
		t.Parallel()

		initSuite := func(suite *WhenRequestingOpenAICompletionsStandardChatCompletionTestingSuite) {
			suite.mockOpenAICompletionsClient.
				EXPECT().
				ChatCompletion(baseUrl, apiKey, gomock.Any()).
				DoAndReturn(func(_ string, _ string, request *external_models.ExternalOpenAICompletionsStandardRequest) (*external_models.ExternalOpenAICompletionsStandardResponse, error) {
					assert.Equal(t, llmConnection.Model, request.Model)
					if assert.Len(t, request.Messages, 1) {
						assert.Equal(t, "user", request.Messages[0].Role)
						if assert.NotNil(t, request.Messages[0].Content) {
							assert.Equal(t, "hello", *request.Messages[0].Content)
						}
					}
					assert.Nil(t, request.Tools)

					return &external_models.ExternalOpenAICompletionsStandardResponse{
						Choices: []external_models.ExternalOpenAICompletionsStandardResponseChoice{
							{
								Message: external_models.ExternalOpenAICompletionsStandardResponseChoiceMessage{
									Role:    "assistant",
									Content: expectedContent,
								},
							},
						},
					}, nil
				})
		}

		t.Run("Should return the assistant text", func(t *testing.T) {
			t.Parallel()

			suite := WhenRequestingOpenAICompletionsStandardChatCompletionBeforeEach(t)
			initSuite(suite)

			result, err := suite.sut.ChatCompletion(baseUrl, apiKey, llmConnection, nil, nil, messages, nil)
			assert.NoError(t, err)
			if assert.NotNil(t, result) {
				assert.Equal(t, []string{expectedContent}, result.Texts)
			}
			assert.Empty(t, result.ToolCalls)
		})
	})

	t.Run("Given the LLM emits tool calls", func(t *testing.T) {
		t.Parallel()

		mcpConnection := model_builders.NewMCPConnectionBuilder().
			WithName("inventory").
			WithUrl("https://mcp.example.com").
			Build()
		mcpTool := &models.Tool{
			Name:        "fetch_items",
			Description: "Fetch items",
			InputScheme: models.ToolInputScheme{
				Type:                 "object",
				Properties:           map[string]models.ToolInputSchemeProperty{},
				Required:             []string{},
				AdditionalProperties: false,
			},
		}
		functionName := fmt.Sprintf("mcp_%s_%s", mcpConnection.Name, mcpTool.Name)
		toolArguments := map[string]any{"query": "value"}
		toolArgumentsJson, _ := json.Marshal(toolArguments)

		mcpTools := map[*models.MCPConnection][]*models.Tool{
			mcpConnection: {mcpTool},
		}

		initSuite := func(suite *WhenRequestingOpenAICompletionsStandardChatCompletionTestingSuite) {
			suite.mockOpenAICompletionsClient.
				EXPECT().
				ChatCompletion(baseUrl, apiKey, gomock.Any()).
				DoAndReturn(func(_ string, _ string, request *external_models.ExternalOpenAICompletionsStandardRequest) (*external_models.ExternalOpenAICompletionsStandardResponse, error) {
					if assert.Len(t, request.Tools, 1) {
						assert.Equal(t, functionName, request.Tools[0].Function.Name)
					}

					return &external_models.ExternalOpenAICompletionsStandardResponse{
						Choices: []external_models.ExternalOpenAICompletionsStandardResponseChoice{
							{
								Message: external_models.ExternalOpenAICompletionsStandardResponseChoiceMessage{
									Role: "assistant",
									ToolCalls: []external_models.ExternalOpenAICompletionsStandardResponseChoiceMessageToolCall{
										{
											ID: "call_1",
											Function: external_models.ExternalOpenAICompletionsStandardToolCallFunction{
												Name:      functionName,
												Arguments: string(toolArgumentsJson),
											},
										},
									},
								},
							},
						},
					}, nil
				})
		}

		t.Run("Should return them in LLMCompletion without resolving them", func(t *testing.T) {
			t.Parallel()

			suite := WhenRequestingOpenAICompletionsStandardChatCompletionBeforeEach(t)
			initSuite(suite)

			result, err := suite.sut.ChatCompletion(baseUrl, apiKey, llmConnection, nil, mcpTools, messages, nil)

			assert.NoError(t, err)
			if assert.NotNil(t, result) && assert.Len(t, result.ToolCalls, 1) {
				assert.Equal(t, "call_1", result.ToolCalls[0].Id)
				assert.Equal(t, functionName, result.ToolCalls[0].Name)
				assert.Equal(t, toolArguments, result.ToolCalls[0].Arguments)
			}
			assert.Empty(t, result.Texts)
		})
	})
}
