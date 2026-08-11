package services_test

import (
	"context"
	"errors"
	"fmt"
	"testing"

	"github.com/6022-labs/agentic-agent-node/src/agent_reasoning/models"
	"github.com/6022-labs/agentic-agent-node/src/agent_reasoning/services"
	"github.com/6022-labs/agentic-agent-node/src/agent_reasoning/services/interfaces"
	"github.com/6022-labs/agentic-agent-node/src/agent_reasoning/settings"
	traces_interfaces "github.com/6022-labs/agentic-agent-node/src/agent_reasoning/traces/interfaces"
	apperrors "github.com/6022-labs/agentic-agent-node/src/common/errors"
	metrics_interfaces_mocks "github.com/6022-labs/agentic-agent-node/tests/agent_reasoning_mocks/metrics_mocks/interfaces_mocks"
	"github.com/6022-labs/agentic-agent-node/tests/agent_reasoning_mocks/services_mocks/interfaces_mocks"
	traces_interfaces_mocks "github.com/6022-labs/agentic-agent-node/tests/agent_reasoning_mocks/traces_mocks/interfaces_mocks"
	"github.com/6022-labs/agentic-agent-node/tests/common_tests/model_builders"
	runtime_services_mocks "github.com/6022-labs/agentic-agent-node/tests/runtime_common_mocks/services_mocks/interfaces_mocks"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
	"go.uber.org/zap"
)

type WhenRunningInferenceLoopTestingSuite struct {
	sut *services.InferenceLoop

	mockLLMExecutor      *interfaces_mocks.MockLLMExecutorInterface
	mockMcpRequester     *interfaces_mocks.MockMcpRequesterInterface
	mockSelfToolRegistry *interfaces_mocks.MockSelfToolRegistryInterface
	mockEmojiDecoder     *runtime_services_mocks.MockEmojiDecoderInterface
}

func WhenRunningInferenceLoopBeforeEach(t *testing.T, maxIterations int) *WhenRunningInferenceLoopTestingSuite {
	return WhenRunningInferenceLoopWithSettingsBeforeEach(
		t,
		&settings.InferenceLoopSettings{
			MaxIterations:      maxIterations,
			MaxDurationSeconds: 300,
			MaxTotalTokens:     1_000_000,
		})
}

func WhenRunningInferenceLoopWithSettingsBeforeEach(t *testing.T, inferenceLoopSettings *settings.InferenceLoopSettings) *WhenRunningInferenceLoopTestingSuite {
	mockController := gomock.NewController(t)

	mockLLMExecutor := interfaces_mocks.NewMockLLMExecutorInterface(mockController)
	mockMcpRequester := interfaces_mocks.NewMockMcpRequesterInterface(mockController)
	mockSelfToolRegistry := interfaces_mocks.NewMockSelfToolRegistryInterface(mockController)
	mockEmojiDecoder := runtime_services_mocks.NewMockEmojiDecoderInterface(mockController)

	mockMetrics := metrics_interfaces_mocks.NewMockReasoningMetricsInterface(mockController)
	mockMetrics.EXPECT().RecordLoopCompleted(gomock.Any(), gomock.Any(), gomock.Any()).AnyTimes()
	mockMetrics.EXPECT().RecordLoopExceeded(gomock.Any(), gomock.Any(), gomock.Any()).AnyTimes()
	mockMetrics.EXPECT().RecordLoopAborted(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).AnyTimes()
	mockMetrics.EXPECT().RecordToolCalls(gomock.Any(), gomock.Any()).AnyTimes()

	mockSpan := traces_interfaces_mocks.NewMockSpan(mockController)
	mockSpan.EXPECT().Fail(gomock.Any()).AnyTimes()
	mockSpan.EXPECT().End().AnyTimes()

	mockReasoningTracer := traces_interfaces_mocks.NewMockReasoningTracerInterface(mockController)
	mockReasoningTracer.EXPECT().StartToolCall(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
		DoAndReturn(func(ctx context.Context, _, _, _ string) (context.Context, traces_interfaces.Span) {
			return ctx, mockSpan
		}).AnyTimes()

	sut := services.NewInferenceLoop(
		zap.NewNop(),
		mockLLMExecutor,
		mockMcpRequester,
		mockSelfToolRegistry,
		inferenceLoopSettings,
		mockMetrics,
		mockReasoningTracer,
		mockEmojiDecoder,
	)

	return &WhenRunningInferenceLoopTestingSuite{
		sut:                  sut,
		mockLLMExecutor:      mockLLMExecutor,
		mockMcpRequester:     mockMcpRequester,
		mockSelfToolRegistry: mockSelfToolRegistry,
		mockEmojiDecoder:     mockEmojiDecoder,
	}
}

func (suite *WhenRunningInferenceLoopTestingSuite) expectDemojizePassthrough() {
	suite.mockEmojiDecoder.EXPECT().DemojizeContent(gomock.Any()).
		DoAndReturn(func(content string) string { return content }).
		AnyTimes()
}

func TestWhenRunningInferenceLoop(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	llmConnection := model_builders.NewLLMConnectionBuilder().Build()

	t.Run("Given the LLM returns final text on the first turn", func(t *testing.T) {
		t.Parallel()

		messages := []*interfaces.ChatCompletionMessage{{Role: "user", Content: "hello"}}
		expectedText := "final answer"

		initSuite := func(suite *WhenRunningInferenceLoopTestingSuite) {
			suite.expectDemojizePassthrough()
			suite.mockLLMExecutor.EXPECT().
				Complete(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), messages).
				Return(&interfaces.LLMCompletion{Texts: []string{expectedText}}, nil)
		}

		t.Run("Should append a single assistant message and stop", func(t *testing.T) {
			t.Parallel()

			suite := WhenRunningInferenceLoopBeforeEach(t, 5)
			initSuite(suite)

			out, err := suite.sut.Run(ctx, llmConnection, nil, nil, messages)

			assert.NoError(t, err)
			if assert.Len(t, out, 2) {
				assert.Equal(t, "user", out[0].Role)
				assert.Equal(t, "assistant", out[1].Role)
				assert.Equal(t, expectedText, out[1].Content)
				assert.Empty(t, out[1].ToolCalls)
			}
		})
	})

	t.Run("Given the completion carries a preamble text block before the final answer", func(t *testing.T) {
		t.Parallel()

		messages := []*interfaces.ChatCompletionMessage{{Role: "user", Content: "hello"}}

		initSuite := func(suite *WhenRunningInferenceLoopTestingSuite) {
			suite.expectDemojizePassthrough()
			suite.mockLLMExecutor.EXPECT().
				Complete(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), messages).
				Return(&interfaces.LLMCompletion{Texts: []string{"preamble", "answer"}}, nil)
		}

		t.Run("Should surface only the last text block as the assistant content", func(t *testing.T) {
			t.Parallel()

			suite := WhenRunningInferenceLoopBeforeEach(t, 5)
			initSuite(suite)

			out, err := suite.sut.Run(ctx, llmConnection, nil, nil, messages)

			assert.NoError(t, err)
			if assert.Len(t, out, 2) {
				assert.Equal(t, "assistant", out[1].Role)
				assert.Equal(t, "answer", out[1].Content)
			}
		})
	})

	t.Run("Given the LLM returns text containing unicode emoji", func(t *testing.T) {
		t.Parallel()

		messages := []*interfaces.ChatCompletionMessage{{Role: "user", Content: "hello"}}
		emojiText := "🔄 status update"

		initSuite := func(suite *WhenRunningInferenceLoopTestingSuite) {
			suite.mockEmojiDecoder.EXPECT().
				DemojizeContent(emojiText).
				Return(":arrows_counterclockwise: status update")
			suite.mockLLMExecutor.EXPECT().
				Complete(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), messages).
				Return(&interfaces.LLMCompletion{Texts: []string{emojiText}}, nil)
		}

		t.Run("Should demojize the assistant content", func(t *testing.T) {
			t.Parallel()

			suite := WhenRunningInferenceLoopBeforeEach(t, 5)
			initSuite(suite)

			out, err := suite.sut.Run(ctx, llmConnection, nil, nil, messages)

			assert.NoError(t, err)
			if assert.Len(t, out, 2) {
				assert.Equal(t, ":arrows_counterclockwise: status update", out[1].Content)
			}
		})
	})

	t.Run("Given the LLM emits an MCP tool call then a final answer", func(t *testing.T) {
		t.Parallel()

		mcpConnection := model_builders.NewMCPConnectionBuilder().
			WithName("inventory").
			Build()

		mcpTool := &models.Tool{Name: "fetch_items"}
		functionName := fmt.Sprintf("mcp_%s_%s", mcpConnection.Name, mcpTool.Name)

		toolArgs := map[string]any{"query": "value"}
		toolResult := map[string]any{"status": "ok"}
		expectedFinal := "all done"

		mcpTools := map[*models.MCPConnection][]*models.Tool{mcpConnection: {mcpTool}}

		initSuite := func(suite *WhenRunningInferenceLoopTestingSuite) {
			suite.expectDemojizePassthrough()
			suite.mockMcpRequester.EXPECT().
				CallTool(gomock.Any(), mcpConnection, mcpTool, toolArgs).
				Return(toolResult, nil)

			firstCall := suite.mockLLMExecutor.EXPECT().
				Complete(gomock.Any(), gomock.Any(), gomock.Any(), mcpTools, gomock.Any()).
				Return(&interfaces.LLMCompletion{
					ToolCalls: []interfaces.ToolCall{
						{Id: "call_1", Name: functionName, Arguments: toolArgs},
					},
				}, nil)

			secondCall := suite.mockLLMExecutor.EXPECT().
				Complete(gomock.Any(), gomock.Any(), gomock.Any(), mcpTools, gomock.Any()).
				DoAndReturn(func(_ context.Context, _ *models.LLMConnection, _ []*models.Tool, _ map[*models.MCPConnection][]*models.Tool, msgs []*interfaces.ChatCompletionMessage) (*interfaces.LLMCompletion, error) {
					if assert.GreaterOrEqual(t, len(msgs), 3) {
						assistant := msgs[len(msgs)-2]
						tool := msgs[len(msgs)-1]
						assert.Equal(t, "assistant", assistant.Role)
						assert.Len(t, assistant.ToolCalls, 1)
						assert.Equal(t, "tool", tool.Role)
						assert.Equal(t, "call_1", tool.ToolCallId)
						assert.Contains(t, tool.Content, "ok")
					}
					return &interfaces.LLMCompletion{Texts: []string{expectedFinal}}, nil
				})

			gomock.InOrder(firstCall, secondCall)
		}

		t.Run("Should resolve the tool, append messages, and re-call until final", func(t *testing.T) {
			t.Parallel()

			suite := WhenRunningInferenceLoopBeforeEach(t, 5)
			initSuite(suite)

			messages := []*interfaces.ChatCompletionMessage{{Role: "user", Content: "go"}}
			out, err := suite.sut.Run(ctx, llmConnection, nil, mcpTools, messages)

			assert.NoError(t, err)
			if assert.Len(t, out, 4) {
				assert.Equal(t, "user", out[0].Role)
				assert.Equal(t, "assistant", out[1].Role)
				assert.Len(t, out[1].ToolCalls, 1)
				assert.Equal(t, "tool", out[2].Role)
				assert.Equal(t, "assistant", out[3].Role)
				assert.Equal(t, expectedFinal, out[3].Content)
			}
		})
	})

	t.Run("Given the LLM emits a self tool call then a final answer", func(t *testing.T) {
		t.Parallel()

		selfTool := &models.Tool{Name: "get_identity"}
		selfTools := []*models.Tool{selfTool}
		toolResult := map[string]any{"name": "agent-x"}
		expectedFinal := "done"

		initSuite := func(suite *WhenRunningInferenceLoopTestingSuite) {
			suite.expectDemojizePassthrough()
			// mockMcpRequester.CallTool is never expected — an MCP dispatch would fail the test.
			suite.mockSelfToolRegistry.EXPECT().
				CallTool(gomock.Any(), selfTool.Name, gomock.Nil()).
				Return(toolResult, nil)

			firstCall := suite.mockLLMExecutor.EXPECT().
				Complete(gomock.Any(), gomock.Any(), selfTools, gomock.Any(), gomock.Any()).
				Return(&interfaces.LLMCompletion{
					ToolCalls: []interfaces.ToolCall{
						{Id: "call_1", Name: "self_get_identity"},
					},
				}, nil)

			secondCall := suite.mockLLMExecutor.EXPECT().
				Complete(gomock.Any(), gomock.Any(), selfTools, gomock.Any(), gomock.Any()).
				Return(&interfaces.LLMCompletion{Texts: []string{expectedFinal}}, nil)

			gomock.InOrder(firstCall, secondCall)
		}

		t.Run("Should dispatch to the self tool registry instead of the MCP requester", func(t *testing.T) {
			t.Parallel()

			suite := WhenRunningInferenceLoopBeforeEach(t, 5)
			initSuite(suite)

			out, err := suite.sut.Run(ctx, llmConnection, selfTools, nil, []*interfaces.ChatCompletionMessage{{Role: "user", Content: "go"}})

			assert.NoError(t, err)
			if assert.Len(t, out, 4) {
				assert.Equal(t, "tool", out[2].Role)
				assert.Equal(t, "call_1", out[2].ToolCallId)
				assert.Contains(t, out[2].Content, "agent-x")
			}
		})
	})

	t.Run("Given the LLM calls a self tool that is not offered this turn", func(t *testing.T) {
		t.Parallel()

		offeredTool := &models.Tool{Name: "get_identity"}
		selfTools := []*models.Tool{offeredTool}
		expectedFinal := "understood"

		initSuite := func(suite *WhenRunningInferenceLoopTestingSuite) {
			suite.expectDemojizePassthrough()
			// mockSelfToolRegistry.CallTool carries no expectation — dispatching the refused tool would fail the test.
			firstCall := suite.mockLLMExecutor.EXPECT().
				Complete(gomock.Any(), gomock.Any(), selfTools, gomock.Any(), gomock.Any()).
				Return(&interfaces.LLMCompletion{
					ToolCalls: []interfaces.ToolCall{
						{Id: "call_1", Name: "self_run_command"},
					},
				}, nil)

			secondCall := suite.mockLLMExecutor.EXPECT().
				Complete(gomock.Any(), gomock.Any(), selfTools, gomock.Any(), gomock.Any()).
				Return(&interfaces.LLMCompletion{Texts: []string{expectedFinal}}, nil)

			gomock.InOrder(firstCall, secondCall)
		}

		t.Run("Should feed a refusal payload back without invoking the registry", func(t *testing.T) {
			t.Parallel()

			suite := WhenRunningInferenceLoopBeforeEach(t, 5)
			initSuite(suite)

			out, err := suite.sut.Run(ctx, llmConnection, selfTools, nil, []*interfaces.ChatCompletionMessage{{Role: "user", Content: "go"}})

			assert.NoError(t, err)
			if assert.Len(t, out, 4) {
				assert.Equal(t, "tool", out[2].Role)
				assert.Equal(t, "call_1", out[2].ToolCallId)
				assert.Contains(t, out[2].Content, "disabled or unknown")
				assert.Equal(t, expectedFinal, out[3].Content)
			}
		})
	})

	t.Run("Given the self tool call fails with a structured validation error", func(t *testing.T) {
		t.Parallel()

		selfTool := &models.Tool{Name: "set_guidance"}
		selfTools := []*models.Tool{selfTool}
		expectedFinal := "understood"

		initSuite := func(suite *WhenRunningInferenceLoopTestingSuite) {
			suite.expectDemojizePassthrough()
			suite.mockSelfToolRegistry.EXPECT().
				CallTool(gomock.Any(), selfTool.Name, gomock.Nil()).
				Return(nil, apperrors.NewValidationError("content", "cannot be empty"))

			firstCall := suite.mockLLMExecutor.EXPECT().
				Complete(gomock.Any(), gomock.Any(), selfTools, gomock.Any(), gomock.Any()).
				Return(&interfaces.LLMCompletion{
					ToolCalls: []interfaces.ToolCall{
						{Id: "call_1", Name: "self_set_guidance"},
					},
				}, nil)

			secondCall := suite.mockLLMExecutor.EXPECT().
				Complete(gomock.Any(), gomock.Any(), selfTools, gomock.Any(), gomock.Any()).
				Return(&interfaces.LLMCompletion{Texts: []string{expectedFinal}}, nil)

			gomock.InOrder(firstCall, secondCall)
		}

		t.Run("Should feed the structured error back as a tool message and continue", func(t *testing.T) {
			t.Parallel()

			suite := WhenRunningInferenceLoopBeforeEach(t, 5)
			initSuite(suite)

			out, err := suite.sut.Run(ctx, llmConnection, selfTools, nil, []*interfaces.ChatCompletionMessage{{Role: "user", Content: "go"}})

			assert.NoError(t, err)
			if assert.Len(t, out, 4) {
				assert.Equal(t, "tool", out[2].Role)
				// The error payload keeps the typed field + message.
				assert.Contains(t, out[2].Content, "\"field\":\"content\"")
				assert.Contains(t, out[2].Content, "cannot be empty")
				assert.Equal(t, expectedFinal, out[3].Content)
			}
		})
	})

	t.Run("Given the self tool call fails with an internal error", func(t *testing.T) {
		t.Parallel()

		selfTool := &models.Tool{Name: "list_mcp_connections"}
		selfTools := []*models.Tool{selfTool}
		expectedFinal := "noted"

		initSuite := func(suite *WhenRunningInferenceLoopTestingSuite) {
			suite.expectDemojizePassthrough()
			suite.mockSelfToolRegistry.EXPECT().
				CallTool(gomock.Any(), selfTool.Name, gomock.Nil()).
				Return(nil, apperrors.NewInternalError("mcp_lookup_failed", "dial tcp 10.0.0.5:5432: connection refused"))

			firstCall := suite.mockLLMExecutor.EXPECT().
				Complete(gomock.Any(), gomock.Any(), selfTools, gomock.Any(), gomock.Any()).
				Return(&interfaces.LLMCompletion{
					ToolCalls: []interfaces.ToolCall{
						{Id: "call_1", Name: "self_list_mcp_connections"},
					},
				}, nil)

			secondCall := suite.mockLLMExecutor.EXPECT().
				Complete(gomock.Any(), gomock.Any(), selfTools, gomock.Any(), gomock.Any()).
				Return(&interfaces.LLMCompletion{Texts: []string{expectedFinal}}, nil)

			gomock.InOrder(firstCall, secondCall)
		}

		t.Run("Should surface the code but hide the infra detail", func(t *testing.T) {
			t.Parallel()

			suite := WhenRunningInferenceLoopBeforeEach(t, 5)
			initSuite(suite)

			out, err := suite.sut.Run(ctx, llmConnection, selfTools, nil, []*interfaces.ChatCompletionMessage{{Role: "user", Content: "go"}})

			assert.NoError(t, err)
			if assert.Len(t, out, 4) {
				assert.Equal(t, "tool", out[2].Role)
				assert.Contains(t, out[2].Content, "mcp_lookup_failed")     // safe code passes through
				assert.Contains(t, out[2].Content, "internal error")        // generic message
				assert.NotContains(t, out[2].Content, "connection refused") // raw infra detail hidden
				assert.NotContains(t, out[2].Content, "5432")
				assert.Equal(t, expectedFinal, out[3].Content)
			}
		})
	})

	t.Run("Given the LLM emits a self tool call and an MCP tool call in the same turn", func(t *testing.T) {
		t.Parallel()

		mcpConnection := model_builders.NewMCPConnectionBuilder().
			WithName("inventory").
			Build()

		selfTool := &models.Tool{Name: "get_identity"}
		selfTools := []*models.Tool{selfTool}
		mcpTool := &models.Tool{Name: "fetch_items"}
		expectedFinal := "done"

		mcpTools := map[*models.MCPConnection][]*models.Tool{
			mcpConnection: {mcpTool},
		}

		initSuite := func(suite *WhenRunningInferenceLoopTestingSuite) {
			suite.expectDemojizePassthrough()
			suite.mockSelfToolRegistry.EXPECT().
				CallTool(gomock.Any(), selfTool.Name, gomock.Nil()).
				Return(map[string]any{"name": "agent-x"}, nil)

			suite.mockMcpRequester.EXPECT().
				CallTool(gomock.Any(), mcpConnection, mcpTool, gomock.Nil()).
				Return(map[string]any{"status": "ok"}, nil)

			firstCall := suite.mockLLMExecutor.EXPECT().
				Complete(gomock.Any(), gomock.Any(), selfTools, mcpTools, gomock.Any()).
				Return(&interfaces.LLMCompletion{
					ToolCalls: []interfaces.ToolCall{
						{Id: "call_1", Name: "self_get_identity"},
						{Id: "call_2", Name: fmt.Sprintf("mcp_%s_%s", mcpConnection.Name, mcpTool.Name)},
					},
				}, nil)

			secondCall := suite.mockLLMExecutor.EXPECT().
				Complete(gomock.Any(), gomock.Any(), selfTools, mcpTools, gomock.Any()).
				Return(&interfaces.LLMCompletion{Texts: []string{expectedFinal}}, nil)

			gomock.InOrder(firstCall, secondCall)
		}

		t.Run("Should dispatch each call to its own path", func(t *testing.T) {
			t.Parallel()

			suite := WhenRunningInferenceLoopBeforeEach(t, 5)
			initSuite(suite)

			out, err := suite.sut.Run(ctx, llmConnection, selfTools, mcpTools, []*interfaces.ChatCompletionMessage{{Role: "user", Content: "go"}})

			assert.NoError(t, err)
			if assert.Len(t, out, 5) {
				assert.Equal(t, "tool", out[2].Role)
				assert.Equal(t, "tool", out[3].Role)
				assert.Equal(t, expectedFinal, out[4].Content)
			}
		})
	})

	t.Run("Given the LLM emits text and tool calls in the same turn", func(t *testing.T) {
		t.Parallel()

		mcpConnection := model_builders.NewMCPConnectionBuilder().
			WithName("inventory").
			Build()

		mcpTool := &models.Tool{Name: "fetch_items"}
		functionName := fmt.Sprintf("mcp_%s_%s", mcpConnection.Name, mcpTool.Name)
		toolArgs := map[string]any{"query": "value"}
		thinkingText := "let me check"
		expectedFinal := "all done"

		mcpTools := map[*models.MCPConnection][]*models.Tool{mcpConnection: {mcpTool}}

		initSuite := func(suite *WhenRunningInferenceLoopTestingSuite) {
			suite.expectDemojizePassthrough()
			suite.mockMcpRequester.EXPECT().
				CallTool(gomock.Any(), mcpConnection, mcpTool, toolArgs).
				Return(map[string]any{"status": "ok"}, nil)

			firstCall := suite.mockLLMExecutor.EXPECT().
				Complete(gomock.Any(), gomock.Any(), gomock.Any(), mcpTools, gomock.Any()).
				Return(&interfaces.LLMCompletion{
					Texts: []string{thinkingText},
					ToolCalls: []interfaces.ToolCall{
						{Id: "call_1", Name: functionName, Arguments: toolArgs},
					},
				}, nil)

			secondCall := suite.mockLLMExecutor.EXPECT().
				Complete(gomock.Any(), gomock.Any(), gomock.Any(), mcpTools, gomock.Any()).
				Return(&interfaces.LLMCompletion{Texts: []string{expectedFinal}}, nil)

			gomock.InOrder(firstCall, secondCall)
		}

		t.Run("Should keep both Content and ToolCalls on the assistant message and continue", func(t *testing.T) {
			t.Parallel()

			suite := WhenRunningInferenceLoopBeforeEach(t, 5)
			initSuite(suite)

			out, err := suite.sut.Run(ctx, llmConnection, nil, mcpTools, []*interfaces.ChatCompletionMessage{{Role: "user", Content: "go"}})

			assert.NoError(t, err)
			if assert.Len(t, out, 4) {
				thinking := out[1]
				assert.Equal(t, "assistant", thinking.Role)
				assert.Equal(t, thinkingText, thinking.Content)
				assert.Len(t, thinking.ToolCalls, 1)

				assert.Equal(t, "tool", out[2].Role)

				final := out[3]
				assert.Equal(t, "assistant", final.Role)
				assert.Equal(t, expectedFinal, final.Content)
				assert.Empty(t, final.ToolCalls)
			}
		})
	})

	t.Run("Given the LLM returns provider-executed tool calls with a final answer", func(t *testing.T) {
		t.Parallel()

		messages := []*interfaces.ChatCompletionMessage{{Role: "user", Content: "go"}}
		expectedFinal := "done"

		initSuite := func(suite *WhenRunningInferenceLoopTestingSuite) {
			suite.expectDemojizePassthrough()
			suite.mockLLMExecutor.EXPECT().
				Complete(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
				Return(&interfaces.LLMCompletion{
					Texts: []string{expectedFinal},
					ExecutedToolCalls: []interfaces.ExecutedToolCall{
						{Id: "mcp_1", Name: "weather_get", Arguments: map[string]any{"city": "paris"}, Output: `{"temp":"20"}`},
					},
				}, nil)
		}

		t.Run("Should record it as assistant+tool turns without dispatching", func(t *testing.T) {
			t.Parallel()

			suite := WhenRunningInferenceLoopBeforeEach(t, 5)
			initSuite(suite)

			// mockMcpRequester.CallTool is never expected — a dispatch would fail the test.
			out, err := suite.sut.Run(ctx, llmConnection, nil, nil, messages)

			assert.NoError(t, err)
			if assert.Len(t, out, 4) {
				assert.Equal(t, "assistant", out[1].Role)
				if assert.Len(t, out[1].ToolCalls, 1) {
					assert.Equal(t, "mcp_1", out[1].ToolCalls[0].Id)
					assert.Equal(t, "weather_get", out[1].ToolCalls[0].Name)
				}
				assert.Equal(t, "tool", out[2].Role)
				assert.Equal(t, "mcp_1", out[2].ToolCallId)
				assert.Equal(t, `{"temp":"20"}`, out[2].Content)
				assert.Equal(t, "assistant", out[3].Role)
				assert.Equal(t, expectedFinal, out[3].Content)
			}
		})
	})

	t.Run("Given the provider runs tools server-side but returns no text", func(t *testing.T) {
		t.Parallel()

		messages := []*interfaces.ChatCompletionMessage{{Role: "user", Content: "go"}}
		finalText := "done"

		initSuite := func(suite *WhenRunningInferenceLoopTestingSuite) {
			suite.expectDemojizePassthrough()
			first := suite.mockLLMExecutor.EXPECT().
				Complete(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
				Return(&interfaces.LLMCompletion{
					ExecutedToolCalls: []interfaces.ExecutedToolCall{
						{Id: "mcp_1", Name: "weather_get", Output: `{"temp":"20"}`},
					},
				}, nil)

			second := suite.mockLLMExecutor.EXPECT().
				Complete(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
				Return(&interfaces.LLMCompletion{Texts: []string{finalText}}, nil)

			gomock.InOrder(first, second)
		}

		t.Run("Should loop again and return the final text", func(t *testing.T) {
			t.Parallel()

			suite := WhenRunningInferenceLoopBeforeEach(t, 5)
			initSuite(suite)

			out, err := suite.sut.Run(ctx, llmConnection, nil, nil, messages)

			assert.NoError(t, err)
			if assert.Len(t, out, 4) {
				assert.Equal(t, "assistant", out[1].Role)
				assert.Len(t, out[1].ToolCalls, 1)
				assert.Equal(t, "tool", out[2].Role)
				assert.Equal(t, "assistant", out[3].Role)
				assert.Equal(t, finalText, out[3].Content)
			}
		})
	})

	t.Run("Given the LLM returns nothing actionable", func(t *testing.T) {
		t.Parallel()

		messages := []*interfaces.ChatCompletionMessage{{Role: "user", Content: "go"}}

		initSuite := func(suite *WhenRunningInferenceLoopTestingSuite) {
			suite.expectDemojizePassthrough()
			suite.mockLLMExecutor.EXPECT().
				Complete(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
				Return(&interfaces.LLMCompletion{}, nil)
		}

		t.Run("Should fail with ErrInferenceLoopNoActionableOutput", func(t *testing.T) {
			t.Parallel()

			suite := WhenRunningInferenceLoopBeforeEach(t, 5)
			initSuite(suite)

			out, err := suite.sut.Run(ctx, llmConnection, nil, nil, messages)

			assert.Nil(t, out)
			assert.ErrorIs(t, err, interfaces.ErrInferenceLoopNoActionableOutput)
		})
	})

	t.Run("Given the LLM keeps requesting tools past max iterations", func(t *testing.T) {
		t.Parallel()

		mcpConnection := model_builders.NewMCPConnectionBuilder().
			WithName("inventory").
			Build()
		mcpTool := &models.Tool{Name: "fetch_items"}
		functionName := fmt.Sprintf("mcp_%s_%s", mcpConnection.Name, mcpTool.Name)
		toolArgs := map[string]any{"query": "value"}

		mcpTools := map[*models.MCPConnection][]*models.Tool{mcpConnection: {mcpTool}}

		initSuite := func(suite *WhenRunningInferenceLoopTestingSuite) {
			suite.expectDemojizePassthrough()
			suite.mockMcpRequester.EXPECT().
				CallTool(gomock.Any(), mcpConnection, mcpTool, toolArgs).
				Return(map[string]any{"status": "ok"}, nil).
				AnyTimes()

			suite.mockLLMExecutor.EXPECT().
				Complete(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
				Return(&interfaces.LLMCompletion{
					ToolCalls: []interfaces.ToolCall{
						{Id: "call_1", Name: functionName, Arguments: toolArgs},
					},
				}, nil).
				AnyTimes()
		}

		t.Run("Should return ErrInferenceLoopMaxIterationsExceeded", func(t *testing.T) {
			t.Parallel()

			suite := WhenRunningInferenceLoopBeforeEach(t, 2)
			initSuite(suite)

			out, err := suite.sut.Run(ctx, llmConnection, nil, mcpTools, nil)

			assert.Nil(t, out)
			assert.ErrorIs(t, err, interfaces.ErrInferenceLoopMaxIterationsExceeded)
		})
	})

	t.Run("Given the LLM calls an MCP tool that is not offered this turn", func(t *testing.T) {
		t.Parallel()

		mcpConnection := model_builders.NewMCPConnectionBuilder().
			WithName("inventory").
			Build()
		offeredTool := &models.Tool{Name: "fetch_items"}
		expectedFinal := "understood"

		mcpTools := map[*models.MCPConnection][]*models.Tool{mcpConnection: {offeredTool}}

		initSuite := func(suite *WhenRunningInferenceLoopTestingSuite) {
			suite.expectDemojizePassthrough()
			// mockMcpRequester.CallTool carries no expectation — dispatching the refused tool would fail the test.
			firstCall := suite.mockLLMExecutor.EXPECT().
				Complete(gomock.Any(), gomock.Any(), gomock.Any(), mcpTools, gomock.Any()).
				Return(&interfaces.LLMCompletion{
					ToolCalls: []interfaces.ToolCall{
						{Id: "call_1", Name: fmt.Sprintf("mcp_%s_ghost_tool", mcpConnection.Name)},
					},
				}, nil)

			secondCall := suite.mockLLMExecutor.EXPECT().
				Complete(gomock.Any(), gomock.Any(), gomock.Any(), mcpTools, gomock.Any()).
				Return(&interfaces.LLMCompletion{Texts: []string{expectedFinal}}, nil)

			gomock.InOrder(firstCall, secondCall)
		}

		t.Run("Should feed a refusal payload back without invoking the requester", func(t *testing.T) {
			t.Parallel()

			suite := WhenRunningInferenceLoopBeforeEach(t, 5)
			initSuite(suite)

			out, err := suite.sut.Run(ctx, llmConnection, nil, mcpTools, []*interfaces.ChatCompletionMessage{{Role: "user", Content: "go"}})

			assert.NoError(t, err)
			if assert.Len(t, out, 4) {
				assert.Equal(t, "tool", out[2].Role)
				assert.Equal(t, "call_1", out[2].ToolCallId)
				assert.Contains(t, out[2].Content, "failed to map tool call")
				assert.Equal(t, expectedFinal, out[3].Content)
			}
		})
	})

	t.Run("Given the LLM emits a tool call with an unknown function name", func(t *testing.T) {
		t.Parallel()

		mcpConnection := model_builders.NewMCPConnectionBuilder().
			WithName("inventory").
			Build()
		mcpTool := &models.Tool{Name: "fetch_items"}
		expectedFinal := "recovered"

		mcpTools := map[*models.MCPConnection][]*models.Tool{mcpConnection: {mcpTool}}

		initSuite := func(suite *WhenRunningInferenceLoopTestingSuite) {
			suite.expectDemojizePassthrough()
			firstCall := suite.mockLLMExecutor.EXPECT().
				Complete(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
				Return(&interfaces.LLMCompletion{
					ToolCalls: []interfaces.ToolCall{
						{Id: "call_x", Name: "ghost_tool", Arguments: nil},
					},
				}, nil)

			secondCall := suite.mockLLMExecutor.EXPECT().
				Complete(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
				Return(&interfaces.LLMCompletion{Texts: []string{expectedFinal}}, nil)

			gomock.InOrder(firstCall, secondCall)
		}

		t.Run("Should feed a tool-mapping error back to the model and continue", func(t *testing.T) {
			t.Parallel()

			suite := WhenRunningInferenceLoopBeforeEach(t, 5)
			initSuite(suite)

			out, err := suite.sut.Run(ctx, llmConnection, nil, mcpTools, []*interfaces.ChatCompletionMessage{{Role: "user", Content: "go"}})

			assert.NoError(t, err)
			if assert.Len(t, out, 4) {
				assert.Equal(t, "tool", out[2].Role)
				assert.Equal(t, "call_x", out[2].ToolCallId)
				assert.Contains(t, out[2].Content, "failed to map tool call")
				assert.Equal(t, expectedFinal, out[3].Content)
			}
		})
	})

	t.Run("Given the LLM executor fails", func(t *testing.T) {
		t.Parallel()

		messages := []*interfaces.ChatCompletionMessage{{Role: "user", Content: "hello"}}

		initSuite := func(suite *WhenRunningInferenceLoopTestingSuite) {
			suite.expectDemojizePassthrough()
			suite.mockLLMExecutor.EXPECT().
				Complete(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), messages).
				Return(nil, errors.New("upstream broke"))
		}

		t.Run("Should wrap the error with the iteration index", func(t *testing.T) {
			t.Parallel()

			suite := WhenRunningInferenceLoopBeforeEach(t, 5)
			initSuite(suite)

			out, err := suite.sut.Run(ctx, llmConnection, nil, nil, messages)

			assert.Nil(t, out)
			assert.Error(t, err)
			assert.ErrorContains(t, err, "llm executor failed at iteration 0")
			assert.ErrorContains(t, err, "upstream broke")
		})
	})

	t.Run("Given the MCP tool call fails", func(t *testing.T) {
		t.Parallel()

		mcpConnection := model_builders.NewMCPConnectionBuilder().
			WithName("inventory").
			Build()
		mcpTool := &models.Tool{Name: "fetch_items"}
		functionName := fmt.Sprintf("mcp_%s_%s", mcpConnection.Name, mcpTool.Name)
		toolArgs := map[string]any{"query": "value"}
		expectedFinal := "recovered"

		mcpTools := map[*models.MCPConnection][]*models.Tool{mcpConnection: {mcpTool}}

		initSuite := func(suite *WhenRunningInferenceLoopTestingSuite) {
			suite.expectDemojizePassthrough()
			firstCall := suite.mockLLMExecutor.EXPECT().
				Complete(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
				Return(&interfaces.LLMCompletion{
					ToolCalls: []interfaces.ToolCall{
						{Id: "call_1", Name: functionName, Arguments: toolArgs},
					},
				}, nil)

			secondCall := suite.mockLLMExecutor.EXPECT().
				Complete(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
				Return(&interfaces.LLMCompletion{Texts: []string{expectedFinal}}, nil)

			gomock.InOrder(firstCall, secondCall)

			suite.mockMcpRequester.EXPECT().
				CallTool(gomock.Any(), mcpConnection, mcpTool, toolArgs).
				Return(nil, errors.New("mcp blew up"))
		}

		t.Run("Should feed a sanitized error back to the model and continue", func(t *testing.T) {
			t.Parallel()

			suite := WhenRunningInferenceLoopBeforeEach(t, 5)
			initSuite(suite)

			out, err := suite.sut.Run(ctx, llmConnection, nil, mcpTools, []*interfaces.ChatCompletionMessage{{Role: "user", Content: "go"}})

			assert.NoError(t, err)
			if assert.Len(t, out, 4) {
				assert.Equal(t, "tool", out[2].Role)
				// The raw upstream error never reaches the model — only a generic message.
				assert.NotContains(t, out[2].Content, "mcp blew up")
				assert.Contains(t, out[2].Content, "the tool call failed")
				assert.Equal(t, expectedFinal, out[3].Content)
			}
		})
	})

	t.Run("Given the loop runs with a wall-clock ceiling", func(t *testing.T) {
		t.Parallel()

		messages := []*interfaces.ChatCompletionMessage{{Role: "user", Content: "go"}}

		t.Run("Should hand the executor a deadline-bounded context", func(t *testing.T) {
			t.Parallel()

			suite := WhenRunningInferenceLoopBeforeEach(t, 5)
			suite.expectDemojizePassthrough()

			var hasDeadline bool
			suite.mockLLMExecutor.EXPECT().
				Complete(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
				DoAndReturn(func(callCtx context.Context, _ *models.LLMConnection, _ []*models.Tool, _ map[*models.MCPConnection][]*models.Tool, _ []*interfaces.ChatCompletionMessage) (*interfaces.LLMCompletion, error) {
					_, hasDeadline = callCtx.Deadline()
					return &interfaces.LLMCompletion{Texts: []string{"done"}}, nil
				})

			_, err := suite.sut.Run(ctx, llmConnection, nil, nil, messages)

			assert.NoError(t, err)
			assert.True(t, hasDeadline)
		})
	})

	t.Run("Given cumulative token usage exceeds the budget between iterations", func(t *testing.T) {
		t.Parallel()

		messages := []*interfaces.ChatCompletionMessage{{Role: "user", Content: "go"}}

		initSuite := func(suite *WhenRunningInferenceLoopTestingSuite) {
			// Provider-executed tool calls with no text force a second iteration without dispatching anything.
			suite.mockLLMExecutor.EXPECT().
				Complete(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
				Return(&interfaces.LLMCompletion{
					ExecutedToolCalls: []interfaces.ExecutedToolCall{{Id: "call_1", Name: "search", Output: "{}"}},
					Usage:             &interfaces.LLMUsage{InputTokens: 80, OutputTokens: 30},
				}, nil)
		}

		t.Run("Should abort with the token budget sentinel before the next round-trip", func(t *testing.T) {
			t.Parallel()

			suite := WhenRunningInferenceLoopWithSettingsBeforeEach(t, &settings.InferenceLoopSettings{
				MaxIterations:      5,
				MaxDurationSeconds: 300,
				MaxTotalTokens:     100,
			})
			initSuite(suite)

			out, err := suite.sut.Run(ctx, llmConnection, nil, nil, messages)

			assert.Nil(t, out)
			assert.ErrorIs(t, err, interfaces.ErrInferenceLoopTokenBudgetExceeded)
		})
	})

	t.Run("Given the executor fails because the wall-clock deadline expired", func(t *testing.T) {
		t.Parallel()

		messages := []*interfaces.ChatCompletionMessage{{Role: "user", Content: "go"}}

		initSuite := func(suite *WhenRunningInferenceLoopTestingSuite) {
			suite.mockLLMExecutor.EXPECT().
				Complete(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
				Return(nil, fmt.Errorf("llm call failed: %w", context.DeadlineExceeded))
		}

		t.Run("Should surface the deadline sentinel instead of a generic executor error", func(t *testing.T) {
			t.Parallel()

			suite := WhenRunningInferenceLoopBeforeEach(t, 5)
			initSuite(suite)

			out, err := suite.sut.Run(ctx, llmConnection, nil, nil, messages)

			assert.Nil(t, out)
			assert.ErrorIs(t, err, interfaces.ErrInferenceLoopDeadlineExceeded)
		})
	})

	t.Run("Given the final answer lands past the budget on the same turn", func(t *testing.T) {
		t.Parallel()

		messages := []*interfaces.ChatCompletionMessage{{Role: "user", Content: "go"}}
		expectedText := "final answer"

		initSuite := func(suite *WhenRunningInferenceLoopTestingSuite) {
			suite.expectDemojizePassthrough()
			suite.mockLLMExecutor.EXPECT().
				Complete(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
				Return(&interfaces.LLMCompletion{
					Texts: []string{expectedText},
					Usage: &interfaces.LLMUsage{InputTokens: 900, OutputTokens: 200},
				}, nil)
		}

		t.Run("Should return the answer instead of discarding it", func(t *testing.T) {
			t.Parallel()

			suite := WhenRunningInferenceLoopWithSettingsBeforeEach(t, &settings.InferenceLoopSettings{
				MaxIterations:      5,
				MaxDurationSeconds: 300,
				MaxTotalTokens:     100,
			})
			initSuite(suite)

			out, err := suite.sut.Run(ctx, llmConnection, nil, nil, messages)

			assert.NoError(t, err)
			if assert.Len(t, out, 2) {
				assert.Equal(t, expectedText, out[1].Content)
			}
		})
	})
}
