package services_test

import (
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

type WhenRequestingAmazonBedrockConverseStandardChatCompletionTestingSuite struct {
	sut *services.AmazonBedrockConverseStandardChatCompletionRequester

	mockAmazonBedrockConverseClient *clients_mocks.MockAmazonBedrockConverseStandardClientInterface
}

func WhenRequestingAmazonBedrockConverseStandardChatCompletionBeforeEach(t *testing.T) *WhenRequestingAmazonBedrockConverseStandardChatCompletionTestingSuite {
	mockController := gomock.NewController(t)

	mockAmazonBedrockConverseClient := clients_mocks.NewMockAmazonBedrockConverseStandardClientInterface(mockController)

	sut := services.NewAmazonBedrockConverseStandardChatCompletionRequester(
		zap.NewNop(),
		mockAmazonBedrockConverseClient,
	)

	return &WhenRequestingAmazonBedrockConverseStandardChatCompletionTestingSuite{
		sut:                             sut,
		mockAmazonBedrockConverseClient: mockAmazonBedrockConverseClient,
	}
}

func TestWhenRequestingAmazonBedrockConverseStandardChatCompletion(t *testing.T) {
	t.Parallel()

	const (
		baseUrl = "https://bedrock.example.com"
		apiKey  = "test-key"
	)

	llmConnection := model_builders.NewLLMConnectionBuilder().
		WithModel("amazon.titan-text-g1").
		Build()

	messages := []*interfaces.ChatCompletionMessage{
		{Role: "user", Content: "hello"},
	}

	t.Run("Given the LLM returns plain text", func(t *testing.T) {
		t.Parallel()

		expectedContent := "hi there"

		initSuite := func(suite *WhenRequestingAmazonBedrockConverseStandardChatCompletionTestingSuite) {
			suite.mockAmazonBedrockConverseClient.
				EXPECT().
				Messages(baseUrl, apiKey, gomock.Any()).
				DoAndReturn(func(_ string, _ string, request *external_models.ExternalAmazonBedrockConverseStandardRequest) (*external_models.ExternalAmazonBedrockConverseStandardResponse, error) {
					assert.Equal(t, llmConnection.Model, request.ModelId)
					if assert.Len(t, request.Messages, 1) {
						assert.Equal(t, "user", request.Messages[0].Role)
					}

					last := expectedContent
					return &external_models.ExternalAmazonBedrockConverseStandardResponse{
						Output: external_models.ExternalAmazonBedrockConverseStandardResponseOutput{
							Message: external_models.ExternalAmazonBedrockConverseStandardMessage{
								Role: "assistant",
								Content: []external_models.ExternalAmazonBedrockConverseStandardMessageContent{
									{Text: &last},
								},
							},
						},
					}, nil
				})
		}

		t.Run("Should return the assistant text", func(t *testing.T) {
			t.Parallel()

			suite := WhenRequestingAmazonBedrockConverseStandardChatCompletionBeforeEach(t)
			initSuite(suite)

			result, err := suite.sut.ChatCompletion(baseUrl, apiKey, llmConnection, nil, nil, messages)
			assert.NoError(t, err)
			if assert.NotNil(t, result) {
				assert.Equal(t, []string{expectedContent}, result.Texts)
			}
		})
	})

	t.Run("Given the LLM returns multiple text blocks", func(t *testing.T) {
		t.Parallel()

		initSuite := func(suite *WhenRequestingAmazonBedrockConverseStandardChatCompletionTestingSuite) {
			suite.mockAmazonBedrockConverseClient.
				EXPECT().
				Messages(baseUrl, apiKey, gomock.Any()).
				DoAndReturn(func(_ string, _ string, _ *external_models.ExternalAmazonBedrockConverseStandardRequest) (*external_models.ExternalAmazonBedrockConverseStandardResponse, error) {
					first := "first"
					second := "second"
					return &external_models.ExternalAmazonBedrockConverseStandardResponse{
						Output: external_models.ExternalAmazonBedrockConverseStandardResponseOutput{
							Message: external_models.ExternalAmazonBedrockConverseStandardMessage{
								Role: "assistant",
								Content: []external_models.ExternalAmazonBedrockConverseStandardMessageContent{
									{Text: &first},
									{Text: &second},
								},
							},
						},
					}, nil
				})
		}

		t.Run("Should collect every text block in order", func(t *testing.T) {
			t.Parallel()

			suite := WhenRequestingAmazonBedrockConverseStandardChatCompletionBeforeEach(t)
			initSuite(suite)

			result, err := suite.sut.ChatCompletion(baseUrl, apiKey, llmConnection, nil, nil, messages)
			assert.NoError(t, err)
			if assert.NotNil(t, result) {
				assert.Equal(t, []string{"first", "second"}, result.Texts)
			}
		})
	})

	t.Run("Given the response has no actionable content", func(t *testing.T) {
		t.Parallel()

		stopReason := "max_tokens"

		initSuite := func(suite *WhenRequestingAmazonBedrockConverseStandardChatCompletionTestingSuite) {
			suite.mockAmazonBedrockConverseClient.
				EXPECT().
				Messages(baseUrl, apiKey, gomock.Any()).
				Return(&external_models.ExternalAmazonBedrockConverseStandardResponse{
					Output: external_models.ExternalAmazonBedrockConverseStandardResponseOutput{
						Message: external_models.ExternalAmazonBedrockConverseStandardMessage{
							Role:    "assistant",
							Content: []external_models.ExternalAmazonBedrockConverseStandardMessageContent{},
						},
					},
					StopReason: &stopReason,
				}, nil)
		}

		t.Run("Should return an empty completion and let the loop decide", func(t *testing.T) {
			t.Parallel()

			suite := WhenRequestingAmazonBedrockConverseStandardChatCompletionBeforeEach(t)
			initSuite(suite)

			result, err := suite.sut.ChatCompletion(baseUrl, apiKey, llmConnection, nil, nil, messages)
			assert.NoError(t, err)
			if assert.NotNil(t, result) {
				assert.Empty(t, result.Texts)
				assert.Empty(t, result.ToolCalls)
				assert.Empty(t, result.ExecutedToolCalls)
			}
		})
	})

	t.Run("Given the LLM emits a toolUse", func(t *testing.T) {
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

		initSuite := func(suite *WhenRequestingAmazonBedrockConverseStandardChatCompletionTestingSuite) {
			suite.mockAmazonBedrockConverseClient.
				EXPECT().
				Messages(baseUrl, apiKey, gomock.Any()).
				DoAndReturn(func(_ string, _ string, request *external_models.ExternalAmazonBedrockConverseStandardRequest) (*external_models.ExternalAmazonBedrockConverseStandardResponse, error) {
					if assert.NotNil(t, request.ToolConfig) && assert.Len(t, request.ToolConfig.Tools, 1) {
						assert.Equal(t, functionName, request.ToolConfig.Tools[0].ToolSpec.Name)
					}

					return &external_models.ExternalAmazonBedrockConverseStandardResponse{
						Output: external_models.ExternalAmazonBedrockConverseStandardResponseOutput{
							Message: external_models.ExternalAmazonBedrockConverseStandardMessage{
								Role: "assistant",
								Content: []external_models.ExternalAmazonBedrockConverseStandardMessageContent{
									{
										ToolUse: &external_models.ExternalAmazonBedrockConverseStandardMessageContentToolUse{
											Name:      functionName,
											Input:     toolArgs,
											ToolUseId: "call_1",
										},
									},
								},
							},
						},
					}, nil
				})
		}

		t.Run("Should return the ToolCall in LLMCompletion", func(t *testing.T) {
			t.Parallel()

			suite := WhenRequestingAmazonBedrockConverseStandardChatCompletionBeforeEach(t)
			initSuite(suite)

			result, err := suite.sut.ChatCompletion(baseUrl, apiKey, llmConnection, nil, mcpTools, messages)
			assert.NoError(t, err)
			if assert.NotNil(t, result) && assert.Len(t, result.ToolCalls, 1) {
				assert.Equal(t, "call_1", result.ToolCalls[0].Id)
				assert.Equal(t, functionName, result.ToolCalls[0].Name)
				assert.Equal(t, toolArgs, result.ToolCalls[0].Arguments)
			}
		})
	})
}
