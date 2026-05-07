package services_test

import (
	"context"
	"errors"
	"testing"

	"github.com/6022-labs/agentic-agent-node/src/agent_reasoning/models"
	"github.com/6022-labs/agentic-agent-node/src/agent_reasoning/services"
	"github.com/6022-labs/agentic-agent-node/src/agent_reasoning/services/interfaces"
	"github.com/6022-labs/agentic-agent-node/tests/agent_reasoning_mocks/repositories_mocks"
	"github.com/6022-labs/agentic-agent-node/tests/agent_reasoning_mocks/services_mocks/interfaces_mocks"
	"github.com/6022-labs/agentic-agent-node/tests/common_tests/model_builders"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

type WhenExecutingLLMTestingSuite struct {
	sut *services.LLMExecutor

	mockLLMConnectionRepository *repositories_mocks.MockLLMConnectionRepositoryInterface
	mockChatCompletionRequester *interfaces_mocks.MockChatCompletionRequesterInterface
}

func WhenExecutingLLMBeforeEach(t *testing.T) *WhenExecutingLLMTestingSuite {
	mockController := gomock.NewController(t)

	mockLLMConnectionRepository := repositories_mocks.NewMockLLMConnectionRepositoryInterface(mockController)
	mockChatCompletionRequester := interfaces_mocks.NewMockChatCompletionRequesterInterface(mockController)

	sut := services.NewLLMExecutor(
		mockLLMConnectionRepository,
		mockChatCompletionRequester,
	)

	return &WhenExecutingLLMTestingSuite{
		sut:                         sut,
		mockLLMConnectionRepository: mockLLMConnectionRepository,
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

	t.Run("Given the LLM connection is configured", func(t *testing.T) {
		t.Parallel()

		llmConnection := model_builders.NewLLMConnectionBuilder().Build()
		expectedText := "hi there"

		initSuite := func(suite *WhenExecutingLLMTestingSuite) {
			suite.mockLLMConnectionRepository.EXPECT().
				Get().
				Return(llmConnection, nil)

			suite.mockChatCompletionRequester.EXPECT().
				ChatCompletion(ctx, llmConnection, gomock.Any(), gomock.Any(), messages).
				Return(&interfaces.LLMCompletion{Text: &expectedText}, nil)
		}

		t.Run("Should delegate to the chat completion requester", func(t *testing.T) {
			t.Parallel()

			suite := WhenExecutingLLMBeforeEach(t)
			initSuite(suite)

			result, err := suite.sut.Execute(ctx, nil, nil, messages)

			assert.NoError(t, err)
			if assert.NotNil(t, result) && assert.NotNil(t, result.Text) {
				assert.Equal(t, expectedText, *result.Text)
			}
		})
	})

	t.Run("Given the LLM connection is not configured", func(t *testing.T) {
		t.Parallel()

		initSuite := func(suite *WhenExecutingLLMTestingSuite) {
			suite.mockLLMConnectionRepository.EXPECT().
				Get().
				Return(nil, nil)
		}

		t.Run("Should return ErrLLMConnectionNotConfigured", func(t *testing.T) {
			t.Parallel()

			suite := WhenExecutingLLMBeforeEach(t)
			initSuite(suite)

			result, err := suite.sut.Execute(ctx, nil, nil, messages)

			assert.Nil(t, result)
			assert.ErrorIs(t, err, interfaces.ErrLLMConnectionNotConfigured)
		})
	})

	t.Run("Given the repository fails", func(t *testing.T) {
		t.Parallel()

		initSuite := func(suite *WhenExecutingLLMTestingSuite) {
			suite.mockLLMConnectionRepository.EXPECT().
				Get().
				Return(nil, errors.New("db down"))
		}

		t.Run("Should wrap the error", func(t *testing.T) {
			t.Parallel()

			suite := WhenExecutingLLMBeforeEach(t)
			initSuite(suite)

			result, err := suite.sut.Execute(ctx, nil, nil, messages)

			assert.Nil(t, result)
			assert.Error(t, err)
			assert.ErrorContains(t, err, "failed to load llm connection")
		})
	})

	t.Run("Given the requester fails", func(t *testing.T) {
		t.Parallel()

		llmConnection := model_builders.NewLLMConnectionBuilder().Build()
		mcpConnections := []*models.MCPConnection{
			model_builders.NewMCPConnectionBuilder().Build(),
		}
		toolsByConn := map[uuid.UUID][]*interfaces.McpTool{
			mcpConnections[0].Id: {{Name: "fetch"}},
		}

		initSuite := func(suite *WhenExecutingLLMTestingSuite) {
			suite.mockLLMConnectionRepository.EXPECT().
				Get().
				Return(llmConnection, nil)

			suite.mockChatCompletionRequester.EXPECT().
				ChatCompletion(ctx, llmConnection, mcpConnections, toolsByConn, messages).
				Return(nil, errors.New("upstream broke"))
		}

		t.Run("Should propagate the error verbatim", func(t *testing.T) {
			t.Parallel()

			suite := WhenExecutingLLMBeforeEach(t)
			initSuite(suite)

			result, err := suite.sut.Execute(ctx, mcpConnections, toolsByConn, messages)

			assert.Nil(t, result)
			assert.Error(t, err)
			assert.ErrorContains(t, err, "upstream broke")
		})
	})
}
