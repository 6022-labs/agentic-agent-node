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

type WhenRequestingAnthropicMessagesStandardChatCompletionTestingSuite struct {
	sut *services.AnthropicMessagesStandardChatCompletionRequester

	mockAnthropicMessagesClient *clients_mocks.MockAnthropicMessagesStandardClientInterface
}

func WhenRequestingAnthropicMessagesStandardChatCompletionBeforeEach(
	t *testing.T,
) *WhenRequestingAnthropicMessagesStandardChatCompletionTestingSuite {
	mockController := gomock.NewController(t)

	mockAnthropicMessagesClient := clients_mocks.NewMockAnthropicMessagesStandardClientInterface(mockController)

	sut := services.NewAnthropicMessagesStandardChatCompletionRequester(
		zap.NewNop(),
		mockAnthropicMessagesClient,
	)

	return &WhenRequestingAnthropicMessagesStandardChatCompletionTestingSuite{
		sut:                         sut,
		mockAnthropicMessagesClient: mockAnthropicMessagesClient,
	}
}

func TestWhenRequestingAnthropicMessagesStandardChatCompletion(t *testing.T) {
	t.Parallel()

	const (
		baseUrl = "https://anthropic.example.com"
		apiKey  = "test-key"
	)

	llmConnection := model_builders.NewLLMConnectionBuilder().
		WithModel("claude-3-5-sonnet").
		Build()

	messages := []*interfaces.ChatCompletionMessage{
		{Role: "system", Content: "you are helpful"},
		{Role: "user", Content: "hello"},
	}

	t.Run("Given the LLM returns plain text", func(t *testing.T) {
		t.Parallel()

		expectedContent := "hi there"

		initSuite := func(suite *WhenRequestingAnthropicMessagesStandardChatCompletionTestingSuite) {
			suite.mockAnthropicMessagesClient.
				EXPECT().
				Messages(baseUrl, apiKey, gomock.Any()).
				DoAndReturn(func(_ string, _ string, request *external_models.ExternalAnthropicMessagesStandardRequest) (*external_models.ExternalAnthropicMessagesStandardResponse, error) {
					assert.Equal(t, llmConnection.Model, request.Model)
					if assert.Len(t, request.System, 1) {
						assert.Equal(t, "you are helpful", request.System[0].Text)
					}
					if assert.Len(t, request.Messages, 1) {
						assert.Equal(t, "user", request.Messages[0].Role)
					}
					assert.Nil(t, request.Tools)
					assert.Nil(t, request.McpServers)

					return &external_models.ExternalAnthropicMessagesStandardResponse{
						ExternalAnthropicMessagesStandardMessage: external_models.ExternalAnthropicMessagesStandardMessage{
							Role: "assistant",
							Content: []external_models.ExternalAnthropicMessagesStandardMessageContent{
								{Type: "text", Text: &expectedContent},
							},
						},
					}, nil
				})
		}

		t.Run("Should return the text content", func(t *testing.T) {
			t.Parallel()

			suite := WhenRequestingAnthropicMessagesStandardChatCompletionBeforeEach(t)
			initSuite(suite)

			result, err := suite.sut.ChatCompletion(baseUrl, apiKey, false, llmConnection, nil, nil, messages, nil)
			assert.NoError(t, err)
			if assert.NotNil(t, result) {
				assert.Equal(t, []string{expectedContent}, result.Texts)
			}
			assert.Empty(t, result.ToolCalls)
		})
	})

	t.Run("Given the LLM emits text before and after a server-side tool call", func(t *testing.T) {
		t.Parallel()

		initSuite := func(suite *WhenRequestingAnthropicMessagesStandardChatCompletionTestingSuite) {
			suite.mockAnthropicMessagesClient.
				EXPECT().
				Messages(baseUrl, apiKey, gomock.Any()).
				DoAndReturn(func(_ string, _ string, _ *external_models.ExternalAnthropicMessagesStandardRequest) (*external_models.ExternalAnthropicMessagesStandardResponse, error) {
					first := "first"
					second := "second"
					id := "mcptoolu_1"
					name := "get"
					isError := false
					return &external_models.ExternalAnthropicMessagesStandardResponse{
						ExternalAnthropicMessagesStandardMessage: external_models.ExternalAnthropicMessagesStandardMessage{
							Role: "assistant",
							Content: []external_models.ExternalAnthropicMessagesStandardMessageContent{
								{Type: "text", Text: &first},
								{Type: "mcp_tool_use", Id: &id, Name: &name, Input: map[string]any{}},
								{
									Type:      "mcp_tool_result",
									ToolUseId: &id,
									IsError:   &isError,
									Content:   []map[string]any{{"type": "text", "text": "20C"}},
								},
								{Type: "text", Text: &second},
							},
						},
					}, nil
				})
		}

		t.Run("Should collect every text block in order", func(t *testing.T) {
			t.Parallel()

			suite := WhenRequestingAnthropicMessagesStandardChatCompletionBeforeEach(t)
			initSuite(suite)

			result, err := suite.sut.ChatCompletion(baseUrl, apiKey, false, llmConnection, nil, nil, messages, nil)
			assert.NoError(t, err)
			if assert.NotNil(t, result) {
				assert.Equal(t, []string{"first", "second"}, result.Texts)
			}
		})
	})

	t.Run("Given the response has no actionable content", func(t *testing.T) {
		t.Parallel()

		stopReason := "max_tokens"

		initSuite := func(suite *WhenRequestingAnthropicMessagesStandardChatCompletionTestingSuite) {
			suite.mockAnthropicMessagesClient.
				EXPECT().
				Messages(baseUrl, apiKey, gomock.Any()).
				Return(&external_models.ExternalAnthropicMessagesStandardResponse{
					ExternalAnthropicMessagesStandardMessage: external_models.ExternalAnthropicMessagesStandardMessage{
						Role:       "assistant",
						Content:    []external_models.ExternalAnthropicMessagesStandardMessageContent{},
						StopReason: &stopReason,
					},
				}, nil)
		}

		t.Run("Should return an empty completion and let the loop decide", func(t *testing.T) {
			t.Parallel()

			suite := WhenRequestingAnthropicMessagesStandardChatCompletionBeforeEach(t)
			initSuite(suite)

			result, err := suite.sut.ChatCompletion(baseUrl, apiKey, false, llmConnection, nil, nil, messages, nil)
			assert.NoError(t, err)
			if assert.NotNil(t, result) {
				assert.Empty(t, result.Texts)
				assert.Empty(t, result.ToolCalls)
				assert.Empty(t, result.ExecutedToolCalls)
			}
		})
	})

	t.Run("Given the LLM emits a tool_use", func(t *testing.T) {
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

		initSuite := func(suite *WhenRequestingAnthropicMessagesStandardChatCompletionTestingSuite) {
			suite.mockAnthropicMessagesClient.
				EXPECT().
				Messages(baseUrl, apiKey, gomock.Any()).
				DoAndReturn(func(_ string, _ string, request *external_models.ExternalAnthropicMessagesStandardRequest) (*external_models.ExternalAnthropicMessagesStandardResponse, error) {
					if assert.Len(t, request.Tools, 1) {
						assert.Equal(t, functionName, request.Tools[0].Name)
					}
					assert.Nil(t, request.McpServers)

					id := "tool_1"
					name := functionName
					return &external_models.ExternalAnthropicMessagesStandardResponse{
						ExternalAnthropicMessagesStandardMessage: external_models.ExternalAnthropicMessagesStandardMessage{
							Role: "assistant",
							Content: []external_models.ExternalAnthropicMessagesStandardMessageContent{
								{Type: "tool_use", Id: &id, Name: &name, Input: toolArgs},
							},
						},
					}, nil
				})
		}

		t.Run("Should return the ToolCall in LLMCompletion", func(t *testing.T) {
			t.Parallel()

			suite := WhenRequestingAnthropicMessagesStandardChatCompletionBeforeEach(t)
			initSuite(suite)

			result, err := suite.sut.ChatCompletion(baseUrl, apiKey, false, llmConnection, nil, mcpTools, messages, nil)
			assert.NoError(t, err)
			if assert.NotNil(t, result) && assert.Len(t, result.ToolCalls, 1) {
				assert.Equal(t, "tool_1", result.ToolCalls[0].Id)
				assert.Equal(t, functionName, result.ToolCalls[0].Name)
				assert.Equal(t, toolArgs, result.ToolCalls[0].Arguments)
			}
		})
	})

	t.Run("Given direct MCP access is enabled", func(t *testing.T) {
		t.Parallel()

		mcpConnection := model_builders.NewMCPConnectionBuilder().Build()
		mcpConnection.HeaderAuth = map[string]string{"Authorization": "Bearer token"}

		expectedContent := "ok"

		initSuite := func(suite *WhenRequestingAnthropicMessagesStandardChatCompletionTestingSuite) {
			suite.mockAnthropicMessagesClient.
				EXPECT().
				Messages(baseUrl, apiKey, gomock.Any()).
				DoAndReturn(func(_ string, _ string, request *external_models.ExternalAnthropicMessagesStandardRequest) (*external_models.ExternalAnthropicMessagesStandardResponse, error) {
					if assert.Len(t, request.McpServers, 1) {
						assert.Equal(t, mcpConnection.Url, request.McpServers[0].Url)
					}
					assert.Nil(t, request.Tools)

					return &external_models.ExternalAnthropicMessagesStandardResponse{
						ExternalAnthropicMessagesStandardMessage: external_models.ExternalAnthropicMessagesStandardMessage{
							Role: "assistant",
							Content: []external_models.ExternalAnthropicMessagesStandardMessageContent{
								{Type: "text", Text: &expectedContent},
							},
						},
					}, nil
				})
		}

		t.Run("Should attach mcp_servers without function-call tools", func(t *testing.T) {
			t.Parallel()

			suite := WhenRequestingAnthropicMessagesStandardChatCompletionBeforeEach(t)
			initSuite(suite)

			result, err := suite.sut.ChatCompletion(
				baseUrl,
				apiKey,
				true,
				llmConnection,
				nil,
				map[*models.MCPConnection][]*models.Tool{mcpConnection: nil},
				messages,
				nil,
			)
			assert.NoError(t, err)
			if assert.NotNil(t, result) {
				assert.Equal(t, []string{expectedContent}, result.Texts)
			}
		})
	})

	t.Run("Given direct MCP access is enabled and self tools are provided", func(t *testing.T) {
		t.Parallel()

		selfTools := []*models.Tool{{Name: "get_identity", Description: "read own identity"}}
		expectedContent := "ok"

		initSuite := func(suite *WhenRequestingAnthropicMessagesStandardChatCompletionTestingSuite) {
			suite.mockAnthropicMessagesClient.
				EXPECT().
				Messages(baseUrl, apiKey, gomock.Any()).
				DoAndReturn(func(_ string, _ string, request *external_models.ExternalAnthropicMessagesStandardRequest) (*external_models.ExternalAnthropicMessagesStandardResponse, error) {
					// Self tools must never become an mcp_server, even in direct-MCP mode.
					assert.Nil(t, request.McpServers)
					if assert.Len(t, request.Tools, 1) {
						assert.Equal(t, "self_get_identity", request.Tools[0].Name)
						// A no-arg tool must serialize required as [], not the invalid "required":null.
						encoded, _ := json.Marshal(request.Tools[0].InputSchema)
						assert.Contains(t, string(encoded), `"required":[]`)
					}

					return &external_models.ExternalAnthropicMessagesStandardResponse{
						ExternalAnthropicMessagesStandardMessage: external_models.ExternalAnthropicMessagesStandardMessage{
							Role: "assistant",
							Content: []external_models.ExternalAnthropicMessagesStandardMessageContent{
								{Type: "text", Text: &expectedContent},
							},
						},
					}, nil
				})
		}

		t.Run("Should serialize them as function-call tools only", func(t *testing.T) {
			t.Parallel()

			suite := WhenRequestingAnthropicMessagesStandardChatCompletionBeforeEach(t)
			initSuite(suite)

			result, err := suite.sut.ChatCompletion(baseUrl, apiKey, true, llmConnection, selfTools, nil, messages, nil)
			assert.NoError(t, err)
			assert.NotNil(t, result)
		})
	})

	t.Run("Given the provider executes an mcp_tool_use server-side", func(t *testing.T) {
		t.Parallel()

		mcpConnection := model_builders.NewMCPConnectionBuilder().WithName("weather").Build()
		mcpConnection.HeaderAuth = map[string]string{"Authorization": "Bearer token"}
		toolArgs := map[string]any{"city": "paris"}

		initSuite := func(suite *WhenRequestingAnthropicMessagesStandardChatCompletionTestingSuite, serverName string) {
			suite.mockAnthropicMessagesClient.
				EXPECT().
				Messages(baseUrl, apiKey, gomock.Any()).
				DoAndReturn(func(_ string, _ string, _ *external_models.ExternalAnthropicMessagesStandardRequest) (*external_models.ExternalAnthropicMessagesStandardResponse, error) {
					id := "mcptoolu_1"
					name := "get"
					label := serverName
					finalText := "it's 20"
					isError := false
					return &external_models.ExternalAnthropicMessagesStandardResponse{
						ExternalAnthropicMessagesStandardMessage: external_models.ExternalAnthropicMessagesStandardMessage{
							Role: "assistant",
							Content: []external_models.ExternalAnthropicMessagesStandardMessageContent{
								{Type: "mcp_tool_use", Id: &id, Name: &name, ServerName: &label, Input: toolArgs},
								{
									Type:      "mcp_tool_result",
									ToolUseId: &id,
									IsError:   &isError,
									Content:   []map[string]any{{"type": "text", "text": "20C"}},
								},
								{Type: "text", Text: &finalText},
							},
						},
					}, nil
				})
		}

		t.Run("Should rebuild the local name when the server is a known connection", func(t *testing.T) {
			t.Parallel()

			suite := WhenRequestingAnthropicMessagesStandardChatCompletionBeforeEach(t)
			initSuite(suite, "weather")

			result, err := suite.sut.ChatCompletion(
				baseUrl,
				apiKey,
				true,
				llmConnection,
				nil,
				map[*models.MCPConnection][]*models.Tool{mcpConnection: nil},
				messages,
				nil,
			)
			assert.NoError(t, err)
			if assert.NotNil(t, result) && assert.Len(t, result.ExecutedToolCalls, 1) {
				assert.Equal(t, "mcptoolu_1", result.ExecutedToolCalls[0].Id)
				assert.Equal(t, "mcp_weather_get", result.ExecutedToolCalls[0].Name)
				assert.Equal(t, toolArgs, result.ExecutedToolCalls[0].Arguments)
				assert.Contains(t, result.ExecutedToolCalls[0].Output, "20C")
				assert.False(t, result.ExecutedToolCalls[0].IsError)
			}
		})

		t.Run("Should keep the provider's raw name when the server is unknown", func(t *testing.T) {
			t.Parallel()

			suite := WhenRequestingAnthropicMessagesStandardChatCompletionBeforeEach(t)
			initSuite(suite, "normalized-weather")

			result, err := suite.sut.ChatCompletion(
				baseUrl,
				apiKey,
				true,
				llmConnection,
				nil,
				map[*models.MCPConnection][]*models.Tool{mcpConnection: nil},
				messages,
				nil,
			)
			assert.NoError(t, err)
			if assert.NotNil(t, result) && assert.Len(t, result.ExecutedToolCalls, 1) {
				assert.Equal(t, "get", result.ExecutedToolCalls[0].Name)
			}
		})
	})
}
