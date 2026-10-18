package services_test

import (
	"context"
	"errors"
	"testing"

	"github.com/6022-labs/agentic-agent-node/src/agent_reasoning/models"
	"github.com/6022-labs/agentic-agent-node/src/agent_reasoning/services"
	"github.com/6022-labs/agentic-agent-node/src/agent_reasoning/services/interfaces"
	traces_interfaces "github.com/6022-labs/agentic-agent-node/src/agent_reasoning/traces/interfaces"
	"github.com/6022-labs/agentic-agent-node/tests/agent_reasoning_mocks/services_mocks/interfaces_mocks"
	traces_interfaces_mocks "github.com/6022-labs/agentic-agent-node/tests/agent_reasoning_mocks/traces_mocks/interfaces_mocks"
	"github.com/6022-labs/agentic-agent-node/tests/common_tests/model_builders"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

type WhenExecutingLLMTestingSuite struct {
	sut *services.LLMExecutor

	mockChatCompletionRequester *interfaces_mocks.MockChatCompletionRequesterInterface
}

func WhenExecutingLLMBeforeEach(t *testing.T) *WhenExecutingLLMTestingSuite {
	mockController := gomock.NewController(t)

	mockChatCompletionRequester := interfaces_mocks.NewMockChatCompletionRequesterInterface(mockController)

	mockChatCompletionSpan := traces_interfaces_mocks.NewMockChatCompletionSpan(mockController)
	mockChatCompletionSpan.EXPECT().RecordUsage(gomock.Any(), gomock.Any()).AnyTimes()
	mockChatCompletionSpan.EXPECT().RecordToolCall(gomock.Any(), gomock.Any()).AnyTimes()
	mockChatCompletionSpan.EXPECT().RecordResponse(gomock.Any(), gomock.Any(), gomock.Any()).AnyTimes()
	mockChatCompletionSpan.EXPECT().Fail(gomock.Any()).AnyTimes()
	mockChatCompletionSpan.EXPECT().End().AnyTimes()

	mockReasoningTracer := traces_interfaces_mocks.NewMockReasoningTracerInterface(mockController)
	mockReasoningTracer.EXPECT().StartChatCompletion(gomock.Any(), gomock.Any(), gomock.Any()).
		DoAndReturn(func(ctx context.Context, _, _ string) (context.Context, traces_interfaces.ChatCompletionSpan) {
			return ctx, mockChatCompletionSpan
		}).AnyTimes()

	sut := services.NewLLMExecutor(
		mockChatCompletionRequester,
		mockReasoningTracer,
	)

	return &WhenExecutingLLMTestingSuite{
		sut:                         sut,
		mockChatCompletionRequester: mockChatCompletionRequester,
	}
}

func TestWhenExecutingLLM(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	userText := "hello"
	messages := []*interfaces.ChatCompletionMessage{
		{Role: "user", Content: userText},
	}

	t.Run("Given the requester fails", func(t *testing.T) {
		t.Parallel()

		llmConnection := model_builders.NewLLMConnectionBuilder().Build()
		mcpConnections := []*models.MCPConnection{
			model_builders.NewMCPConnectionBuilder().Build(),
		}
		mcpTools := map[*models.MCPConnection][]*models.Tool{
			mcpConnections[0]: {{Name: "fetch"}},
		}

		initSuite := func(suite *WhenExecutingLLMTestingSuite) {
			suite.mockChatCompletionRequester.EXPECT().
				ChatCompletion(gomock.Any(), llmConnection, gomock.Nil(), mcpTools, messages, gomock.Any()).
				Return(nil, errors.New("upstream broke"))
		}

		t.Run("Should propagate the error verbatim", func(t *testing.T) {
			t.Parallel()

			suite := WhenExecutingLLMBeforeEach(t)
			initSuite(suite)

			result, err := suite.sut.Complete(ctx, llmConnection, nil, mcpTools, messages)

			assert.Nil(t, result)
			assert.ErrorContains(t, err, "upstream broke")
		})
	})

	t.Run("Given a resolved LLM connection", func(t *testing.T) {
		t.Parallel()

		llmConnection := model_builders.NewLLMConnectionBuilder().Build()
		expectedText := "hi there"

		initSuite := func(suite *WhenExecutingLLMTestingSuite) {
			suite.mockChatCompletionRequester.EXPECT().
				ChatCompletion(gomock.Any(), llmConnection, gomock.Any(), gomock.Any(), messages, gomock.Any()).
				Return(&interfaces.LLMCompletion{Texts: []string{expectedText}}, nil)
		}

		t.Run("Should delegate to the chat completion requester", func(t *testing.T) {
			t.Parallel()

			suite := WhenExecutingLLMBeforeEach(t)
			initSuite(suite)

			result, err := suite.sut.Complete(ctx, llmConnection, nil, nil, messages)

			assert.NoError(t, err)
			if assert.NotNil(t, result) {
				assert.Equal(t, []string{expectedText}, result.Texts)
			}
		})
	})
}
