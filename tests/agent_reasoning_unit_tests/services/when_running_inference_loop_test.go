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
	metrics_interfaces_mocks "github.com/6022-labs/agentic-agent-node/tests/agent_reasoning_mocks/metrics_mocks/interfaces_mocks"
	"github.com/6022-labs/agentic-agent-node/tests/agent_reasoning_mocks/services_mocks/interfaces_mocks"
	traces_interfaces_mocks "github.com/6022-labs/agentic-agent-node/tests/agent_reasoning_mocks/traces_mocks/interfaces_mocks"
	"github.com/6022-labs/agentic-agent-node/tests/common_tests/model_builders"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
	"go.uber.org/zap"
)

type WhenRunningInferenceLoopTestingSuite struct {
	sut *services.InferenceLoop

	mockLLMExecutor  *interfaces_mocks.MockLLMExecutorInterface
	mockMcpRequester *interfaces_mocks.MockMcpRequesterInterface
}

func WhenRunningInferenceLoopBeforeEach(t *testing.T, maxIterations int) *WhenRunningInferenceLoopTestingSuite {
	mockController := gomock.NewController(t)

	mockLLMExecutor := interfaces_mocks.NewMockLLMExecutorInterface(mockController)
	mockMcpRequester := interfaces_mocks.NewMockMcpRequesterInterface(mockController)

	mockMetrics := metrics_interfaces_mocks.NewMockReasoningMetricsInterface(mockController)
	mockMetrics.EXPECT().RecordLoopCompleted(gomock.Any(), gomock.Any(), gomock.Any()).AnyTimes()
	mockMetrics.EXPECT().RecordLoopExceeded(gomock.Any(), gomock.Any(), gomock.Any()).AnyTimes()
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
		&settings.InferenceLoopSettings{MaxIterations: maxIterations},
		mockMetrics,
		mockReasoningTracer,
	)

	return &WhenRunningInferenceLoopTestingSuite{
		sut:              sut,
		mockLLMExecutor:  mockLLMExecutor,
		mockMcpRequester: mockMcpRequester,
	}
}

func TestWhenRunningInferenceLoop(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	t.Run("Given the LLM returns text containing unicode emoji", func(t *testing.T) {
		t.Parallel()

		messages := []*interfaces.ChatCompletionMessage{{Role: "user", Content: "hello"}}
		emojiText := "🔄 status update"

		initSuite := func(suite *WhenRunningInferenceLoopTestingSuite) {
			suite.mockLLMExecutor.EXPECT().
				Complete(ctx, gomock.Nil(), gomock.Any(), messages).
				Return(&interfaces.LLMCompletion{Text: &emojiText}, nil)
		}

		t.Run("Should demojize the assistant content", func(t *testing.T) {
			t.Parallel()

			suite := WhenRunningInferenceLoopBeforeEach(t, 5)
			initSuite(suite)

			out, err := suite.sut.Run(ctx, nil, nil, messages)

			assert.NoError(t, err)
			if assert.Len(t, out, 2) {
				assert.Equal(t, ":arrows_counterclockwise: status update", out[1].Content)
			}
		})
	})

	t.Run("Given the LLM returns final text on the first turn", func(t *testing.T) {
		t.Parallel()

		messages := []*interfaces.ChatCompletionMessage{{Role: "user", Content: "hello"}}
		expectedText := "final answer"

		initSuite := func(suite *WhenRunningInferenceLoopTestingSuite) {
			suite.mockLLMExecutor.EXPECT().
				Complete(ctx, gomock.Nil(), gomock.Any(), messages).
				Return(&interfaces.LLMCompletion{Text: &expectedText}, nil)
		}

		t.Run("Should append a single assistant message and stop", func(t *testing.T) {
			t.Parallel()

			suite := WhenRunningInferenceLoopBeforeEach(t, 5)
			initSuite(suite)

			out, err := suite.sut.Run(ctx, nil, nil, messages)

			assert.NoError(t, err)
			if assert.Len(t, out, 2) {
				assert.Equal(t, "user", out[0].Role)
				assert.Equal(t, "assistant", out[1].Role)
				assert.Equal(t, expectedText, out[1].Content)
				assert.Empty(t, out[1].ToolCalls)
			}
		})
	})

	t.Run("Given the LLM emits a tool call then a final answer", func(t *testing.T) {
		t.Parallel()

		mcpConnection := model_builders.NewMCPConnectionBuilder().
			WithName("inventory").
			Build()
		mcpConnections := []*models.MCPConnection{mcpConnection}

		mcpTool := &interfaces.McpTool{Name: "fetch_items"}
		functionName := fmt.Sprintf("%s_%s", mcpConnection.Name, mcpTool.Name)

		toolArgs := map[string]any{"query": "value"}
		toolResult := map[string]any{"status": "ok"}
		expectedFinal := "all done"

		toolsByConnectionId := map[uuid.UUID][]*interfaces.McpTool{mcpConnection.Id: {mcpTool}}

		initSuite := func(suite *WhenRunningInferenceLoopTestingSuite) {
			suite.mockMcpRequester.EXPECT().
				CallTool(mcpConnection, mcpTool, toolArgs).
				Return(toolResult, nil)

			firstCall := suite.mockLLMExecutor.EXPECT().
				Complete(ctx, mcpConnections, gomock.Any(), gomock.Any()).
				Return(&interfaces.LLMCompletion{
					ToolCalls: []interfaces.ToolCall{
						{Id: "call_1", Name: functionName, Arguments: toolArgs},
					},
				}, nil)

			secondCall := suite.mockLLMExecutor.EXPECT().
				Complete(ctx, mcpConnections, gomock.Any(), gomock.Any()).
				DoAndReturn(func(_ context.Context, _ []*models.MCPConnection, _ map[uuid.UUID][]*interfaces.McpTool, msgs []*interfaces.ChatCompletionMessage) (*interfaces.LLMCompletion, error) {
					if assert.GreaterOrEqual(t, len(msgs), 3) {
						assistant := msgs[len(msgs)-2]
						tool := msgs[len(msgs)-1]
						assert.Equal(t, "assistant", assistant.Role)
						assert.Len(t, assistant.ToolCalls, 1)
						assert.Equal(t, "tool", tool.Role)
						assert.Equal(t, "call_1", tool.ToolCallId)
						assert.Contains(t, tool.Content, "ok")
					}
					return &interfaces.LLMCompletion{Text: &expectedFinal}, nil
				})

			gomock.InOrder(firstCall, secondCall)
		}

		t.Run("Should resolve the tool, append messages, and re-call until final", func(t *testing.T) {
			t.Parallel()

			suite := WhenRunningInferenceLoopBeforeEach(t, 5)
			initSuite(suite)

			messages := []*interfaces.ChatCompletionMessage{{Role: "user", Content: "go"}}
			out, err := suite.sut.Run(ctx, mcpConnections, toolsByConnectionId, messages)

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

	t.Run("Given the LLM returns provider-executed tool calls with a final answer", func(t *testing.T) {
		t.Parallel()

		messages := []*interfaces.ChatCompletionMessage{{Role: "user", Content: "go"}}
		expectedFinal := "done"

		initSuite := func(suite *WhenRunningInferenceLoopTestingSuite) {
			suite.mockLLMExecutor.EXPECT().
				Complete(ctx, gomock.Nil(), gomock.Any(), gomock.Any()).
				Return(&interfaces.LLMCompletion{
					Text: &expectedFinal,
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
			out, err := suite.sut.Run(ctx, nil, nil, messages)

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
			first := suite.mockLLMExecutor.EXPECT().
				Complete(ctx, gomock.Nil(), gomock.Any(), gomock.Any()).
				Return(&interfaces.LLMCompletion{
					ExecutedToolCalls: []interfaces.ExecutedToolCall{
						{Id: "mcp_1", Name: "weather_get", Output: `{"temp":"20"}`},
					},
				}, nil)

			second := suite.mockLLMExecutor.EXPECT().
				Complete(ctx, gomock.Nil(), gomock.Any(), gomock.Any()).
				Return(&interfaces.LLMCompletion{Text: &finalText}, nil)

			gomock.InOrder(first, second)
		}

		t.Run("Should loop again and return the final text", func(t *testing.T) {
			t.Parallel()

			suite := WhenRunningInferenceLoopBeforeEach(t, 5)
			initSuite(suite)

			out, err := suite.sut.Run(ctx, nil, nil, messages)

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
			suite.mockLLMExecutor.EXPECT().
				Complete(ctx, gomock.Nil(), gomock.Any(), gomock.Any()).
				Return(&interfaces.LLMCompletion{}, nil)
		}

		t.Run("Should fail with ErrInferenceLoopNoActionableOutput", func(t *testing.T) {
			t.Parallel()

			suite := WhenRunningInferenceLoopBeforeEach(t, 5)
			initSuite(suite)

			out, err := suite.sut.Run(ctx, nil, nil, messages)

			assert.Nil(t, out)
			assert.ErrorIs(t, err, interfaces.ErrInferenceLoopNoActionableOutput)
		})
	})

	t.Run("Given the LLM keeps requesting tools past max iterations", func(t *testing.T) {
		t.Parallel()

		mcpConnection := model_builders.NewMCPConnectionBuilder().
			WithName("inventory").
			Build()
		mcpTool := &interfaces.McpTool{Name: "fetch_items"}
		functionName := fmt.Sprintf("%s_%s", mcpConnection.Name, mcpTool.Name)
		toolArgs := map[string]any{"query": "value"}

		toolsByConnectionId := map[uuid.UUID][]*interfaces.McpTool{mcpConnection.Id: {mcpTool}}

		initSuite := func(suite *WhenRunningInferenceLoopTestingSuite) {
			suite.mockMcpRequester.EXPECT().
				CallTool(mcpConnection, mcpTool, toolArgs).
				Return(map[string]any{"status": "ok"}, nil).
				AnyTimes()

			suite.mockLLMExecutor.EXPECT().
				Complete(ctx, gomock.Any(), gomock.Any(), gomock.Any()).
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

			out, err := suite.sut.Run(ctx, []*models.MCPConnection{mcpConnection}, toolsByConnectionId, nil)

			assert.Nil(t, out)
			assert.ErrorIs(t, err, interfaces.ErrInferenceLoopMaxIterationsExceeded)
		})
	})

	t.Run("Given the LLM emits a tool call with an unknown function name", func(t *testing.T) {
		t.Parallel()

		mcpConnection := model_builders.NewMCPConnectionBuilder().
			WithName("inventory").
			Build()
		mcpTool := &interfaces.McpTool{Name: "fetch_items"}

		toolsByConnectionId := map[uuid.UUID][]*interfaces.McpTool{mcpConnection.Id: {mcpTool}}

		initSuite := func(suite *WhenRunningInferenceLoopTestingSuite) {
			suite.mockLLMExecutor.EXPECT().
				Complete(ctx, gomock.Any(), gomock.Any(), gomock.Any()).
				Return(&interfaces.LLMCompletion{
					ToolCalls: []interfaces.ToolCall{
						{Id: "call_x", Name: "ghost_tool", Arguments: nil},
					},
				}, nil)
		}

		t.Run("Should fail with a tool-mapping error", func(t *testing.T) {
			t.Parallel()

			suite := WhenRunningInferenceLoopBeforeEach(t, 5)
			initSuite(suite)

			out, err := suite.sut.Run(ctx, []*models.MCPConnection{mcpConnection}, toolsByConnectionId, nil)

			assert.Nil(t, out)
			assert.Error(t, err)
			assert.ErrorContains(t, err, "failed to map tool call")
		})
	})

	t.Run("Given the LLM executor fails", func(t *testing.T) {
		t.Parallel()

		messages := []*interfaces.ChatCompletionMessage{{Role: "user", Content: "hello"}}

		initSuite := func(suite *WhenRunningInferenceLoopTestingSuite) {
			suite.mockLLMExecutor.EXPECT().
				Complete(ctx, gomock.Nil(), gomock.Any(), messages).
				Return(nil, errors.New("upstream broke"))
		}

		t.Run("Should wrap the error with the iteration index", func(t *testing.T) {
			t.Parallel()

			suite := WhenRunningInferenceLoopBeforeEach(t, 5)
			initSuite(suite)

			out, err := suite.sut.Run(ctx, nil, nil, messages)

			assert.Nil(t, out)
			assert.Error(t, err)
			assert.ErrorContains(t, err, "llm executor failed at iteration 0")
			assert.ErrorContains(t, err, "upstream broke")
		})
	})

	t.Run("Given the LLM emits text and tool calls in the same turn", func(t *testing.T) {
		t.Parallel()

		mcpConnection := model_builders.NewMCPConnectionBuilder().
			WithName("inventory").
			Build()
		mcpConnections := []*models.MCPConnection{mcpConnection}

		mcpTool := &interfaces.McpTool{Name: "fetch_items"}
		functionName := fmt.Sprintf("%s_%s", mcpConnection.Name, mcpTool.Name)
		toolArgs := map[string]any{"query": "value"}
		thinkingText := "let me check"
		expectedFinal := "all done"

		toolsByConnectionId := map[uuid.UUID][]*interfaces.McpTool{mcpConnection.Id: {mcpTool}}

		initSuite := func(suite *WhenRunningInferenceLoopTestingSuite) {
			suite.mockMcpRequester.EXPECT().
				CallTool(mcpConnection, mcpTool, toolArgs).
				Return(map[string]any{"status": "ok"}, nil)

			firstCall := suite.mockLLMExecutor.EXPECT().
				Complete(ctx, mcpConnections, gomock.Any(), gomock.Any()).
				Return(&interfaces.LLMCompletion{
					Text: &thinkingText,
					ToolCalls: []interfaces.ToolCall{
						{Id: "call_1", Name: functionName, Arguments: toolArgs},
					},
				}, nil)

			secondCall := suite.mockLLMExecutor.EXPECT().
				Complete(ctx, mcpConnections, gomock.Any(), gomock.Any()).
				Return(&interfaces.LLMCompletion{Text: &expectedFinal}, nil)

			gomock.InOrder(firstCall, secondCall)
		}

		t.Run("Should keep both Content and ToolCalls on the assistant message and continue", func(t *testing.T) {
			t.Parallel()

			suite := WhenRunningInferenceLoopBeforeEach(t, 5)
			initSuite(suite)

			out, err := suite.sut.Run(ctx, mcpConnections, toolsByConnectionId, []*interfaces.ChatCompletionMessage{{Role: "user", Content: "go"}})

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

	t.Run("Given the MCP tool call fails", func(t *testing.T) {
		t.Parallel()

		mcpConnection := model_builders.NewMCPConnectionBuilder().
			WithName("inventory").
			Build()
		mcpTool := &interfaces.McpTool{Name: "fetch_items"}
		functionName := fmt.Sprintf("%s_%s", mcpConnection.Name, mcpTool.Name)
		toolArgs := map[string]any{"query": "value"}

		toolsByConnectionId := map[uuid.UUID][]*interfaces.McpTool{mcpConnection.Id: {mcpTool}}

		initSuite := func(suite *WhenRunningInferenceLoopTestingSuite) {
			suite.mockLLMExecutor.EXPECT().
				Complete(ctx, gomock.Any(), gomock.Any(), gomock.Any()).
				Return(&interfaces.LLMCompletion{
					ToolCalls: []interfaces.ToolCall{
						{Id: "call_1", Name: functionName, Arguments: toolArgs},
					},
				}, nil)

			suite.mockMcpRequester.EXPECT().
				CallTool(mcpConnection, mcpTool, toolArgs).
				Return(nil, errors.New("mcp blew up"))
		}

		t.Run("Should fail without re-calling the LLM", func(t *testing.T) {
			t.Parallel()

			suite := WhenRunningInferenceLoopBeforeEach(t, 5)
			initSuite(suite)

			out, err := suite.sut.Run(ctx, []*models.MCPConnection{mcpConnection}, toolsByConnectionId, nil)

			assert.Nil(t, out)
			assert.Error(t, err)
			assert.ErrorContains(t, err, "failed to call mcp tool")
			assert.ErrorContains(t, err, "mcp blew up")
		})
	})
}
