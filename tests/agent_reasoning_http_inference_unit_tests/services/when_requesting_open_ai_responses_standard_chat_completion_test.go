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
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
	"go.uber.org/zap"
)

type WhenRequestingOpenAIResponsesStandardChatCompletionTestingSuite struct {
	sut *services.OpenAIResponsesStandardChatCompletionRequester

	mockOpenAIResponsesClient *clients_mocks.MockOpenAIResponsesStandardClientInterface
}

func WhenRequestingOpenAIResponsesStandardChatCompletionBeforeEach(t *testing.T) *WhenRequestingOpenAIResponsesStandardChatCompletionTestingSuite {
	mockController := gomock.NewController(t)

	mockOpenAIResponsesClient := clients_mocks.NewMockOpenAIResponsesStandardClientInterface(mockController)

	sut := services.NewOpenAIResponsesStandardChatCompletionRequester(
		zap.NewNop(),
		mockOpenAIResponsesClient,
	)

	return &WhenRequestingOpenAIResponsesStandardChatCompletionTestingSuite{
		sut:                       sut,
		mockOpenAIResponsesClient: mockOpenAIResponsesClient,
	}
}

func TestWhenRequestingOpenAIResponsesStandardChatCompletion(t *testing.T) {
	t.Parallel()

	const (
		baseUrl = "https://openai.example.com"
		apiKey  = "test-key"
	)

	llmConnection := model_builders.NewLLMConnectionBuilder().
		WithModel("gpt-4").
		Build()

	messages := []*interfaces.ChatCompletionMessage{
		{Role: "user", Content: "hello"},
	}

	t.Run("Given the LLM returns plain text", func(t *testing.T) {
		t.Parallel()

		expectedContent := "hi there"

		initSuite := func(suite *WhenRequestingOpenAIResponsesStandardChatCompletionTestingSuite) {
			suite.mockOpenAIResponsesClient.
				EXPECT().
				Responses(baseUrl, apiKey, gomock.Any()).
				DoAndReturn(func(_ string, _ string, request *external_models.ExternalOpenAIResponsesStandardRequest) (*external_models.ExternalOpenAIResponsesStandardResponse, error) {
					assert.Equal(t, llmConnection.Model, request.Model)
					if assert.Len(t, request.Input, 1) {
						assert.Equal(t, "message", request.Input[0].Type)
						if assert.NotNil(t, request.Input[0].Role) {
							assert.Equal(t, "user", *request.Input[0].Role)
						}
					}
					assert.Nil(t, request.PreviousResponseID)
					assert.Nil(t, request.Tools)

					assistantRole := "assistant"
					return &external_models.ExternalOpenAIResponsesStandardResponse{
						Output: []external_models.ExternalOpenAIResponsesStandardResponseOutput{
							{
								Type: "message",
								Role: &assistantRole,
								Content: []external_models.ExternalOpenAIResponsesStandardResponseOutputContent{
									{Type: "output_text", Text: expectedContent},
								},
							},
						},
					}, nil
				})
		}

		t.Run("Should return the assistant text", func(t *testing.T) {
			t.Parallel()

			suite := WhenRequestingOpenAIResponsesStandardChatCompletionBeforeEach(t)
			initSuite(suite)

			result, err := suite.sut.ChatCompletion(baseUrl, apiKey, false, llmConnection, nil, nil, messages)
			assert.NoError(t, err)
			if assert.NotNil(t, result) && assert.NotNil(t, result.Text) {
				assert.Equal(t, expectedContent, *result.Text)
			}
		})
	})

	t.Run("Given the response has no actionable output", func(t *testing.T) {
		t.Parallel()

		initSuite := func(suite *WhenRequestingOpenAIResponsesStandardChatCompletionTestingSuite) {
			suite.mockOpenAIResponsesClient.
				EXPECT().
				Responses(baseUrl, apiKey, gomock.Any()).
				Return(&external_models.ExternalOpenAIResponsesStandardResponse{
					Status: "incomplete",
					Output: []external_models.ExternalOpenAIResponsesStandardResponseOutput{},
				}, nil)
		}

		t.Run("Should return an error", func(t *testing.T) {
			t.Parallel()

			suite := WhenRequestingOpenAIResponsesStandardChatCompletionBeforeEach(t)
			initSuite(suite)

			result, err := suite.sut.ChatCompletion(baseUrl, apiKey, false, llmConnection, nil, nil, messages)
			assert.Nil(t, result)
			assert.Error(t, err)
			assert.ErrorContains(t, err, "openai responses returned no actionable output")
			assert.ErrorContains(t, err, "incomplete")
		})
	})

	t.Run("Given the LLM emits a function_call", func(t *testing.T) {
		t.Parallel()

		mcpConnection := model_builders.NewMCPConnectionBuilder().
			WithName("inventory").
			Build()
		mcpTool := &interfaces.McpTool{Name: "fetch_items"}
		functionName := fmt.Sprintf("%s_%s", mcpConnection.Name, mcpTool.Name)
		toolArgs := map[string]any{"query": "value"}
		argsJson, _ := json.Marshal(toolArgs)
		argsStr := string(argsJson)

		toolsByConn := map[uuid.UUID][]*interfaces.McpTool{
			mcpConnection.Id: {mcpTool},
		}

		initSuite := func(suite *WhenRequestingOpenAIResponsesStandardChatCompletionTestingSuite) {
			suite.mockOpenAIResponsesClient.
				EXPECT().
				Responses(baseUrl, apiKey, gomock.Any()).
				DoAndReturn(func(_ string, _ string, request *external_models.ExternalOpenAIResponsesStandardRequest) (*external_models.ExternalOpenAIResponsesStandardResponse, error) {
					if assert.Len(t, request.Tools, 1) {
						assert.Equal(t, "function", request.Tools[0].Type)
						if assert.NotNil(t, request.Tools[0].Name) {
							assert.Equal(t, functionName, *request.Tools[0].Name)
						}
					}

					name := functionName
					callId := "call_1"
					return &external_models.ExternalOpenAIResponsesStandardResponse{
						Output: []external_models.ExternalOpenAIResponsesStandardResponseOutput{
							{
								Type:      "function_call",
								Name:      &name,
								CallId:    &callId,
								Arguments: &argsStr,
							},
						},
					}, nil
				})
		}

		t.Run("Should return the ToolCall in LLMCompletion", func(t *testing.T) {
			t.Parallel()

			suite := WhenRequestingOpenAIResponsesStandardChatCompletionBeforeEach(t)
			initSuite(suite)

			result, err := suite.sut.ChatCompletion(baseUrl, apiKey, false, llmConnection, []*models.MCPConnection{mcpConnection}, toolsByConn, messages)
			assert.NoError(t, err)
			if assert.NotNil(t, result) && assert.Len(t, result.ToolCalls, 1) {
				assert.Equal(t, "call_1", result.ToolCalls[0].Id)
				assert.Equal(t, functionName, result.ToolCalls[0].Name)
				assert.Equal(t, toolArgs, result.ToolCalls[0].Arguments)
			}
		})
	})

	t.Run("Given direct MCP access is enabled", func(t *testing.T) {
		t.Parallel()

		mcpConnection := model_builders.NewMCPConnectionBuilder().Build()
		expectedContent := "ok"

		initSuite := func(suite *WhenRequestingOpenAIResponsesStandardChatCompletionTestingSuite) {
			suite.mockOpenAIResponsesClient.
				EXPECT().
				Responses(baseUrl, apiKey, gomock.Any()).
				DoAndReturn(func(_ string, _ string, request *external_models.ExternalOpenAIResponsesStandardRequest) (*external_models.ExternalOpenAIResponsesStandardResponse, error) {
					if assert.Len(t, request.Tools, 1) {
						assert.Equal(t, "mcp", request.Tools[0].Type)
					}

					assistantRole := "assistant"
					return &external_models.ExternalOpenAIResponsesStandardResponse{
						Output: []external_models.ExternalOpenAIResponsesStandardResponseOutput{
							{
								Type: "message",
								Role: &assistantRole,
								Content: []external_models.ExternalOpenAIResponsesStandardResponseOutputContent{
									{Type: "output_text", Text: expectedContent},
								},
							},
						},
					}, nil
				})
		}

		t.Run("Should attach mcp tools without function-call tools", func(t *testing.T) {
			t.Parallel()

			suite := WhenRequestingOpenAIResponsesStandardChatCompletionBeforeEach(t)
			initSuite(suite)

			result, err := suite.sut.ChatCompletion(baseUrl, apiKey, true, llmConnection, []*models.MCPConnection{mcpConnection}, nil, messages)
			assert.NoError(t, err)
			if assert.NotNil(t, result) && assert.NotNil(t, result.Text) {
				assert.Equal(t, expectedContent, *result.Text)
			}
		})
	})
}
