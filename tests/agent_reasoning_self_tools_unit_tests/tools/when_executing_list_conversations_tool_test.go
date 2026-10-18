package tools_test

import (
	"context"
	"errors"
	"strings"
	"testing"

	knowledge_requests "github.com/6022-labs/agentic-agent-node/src/agent_knowledge/use_cases/requests"
	knowledge_responses "github.com/6022-labs/agentic-agent-node/src/agent_knowledge/use_cases/responses"
	"github.com/6022-labs/agentic-agent-node/src/agent_reasoning_self_tools/tools"
	"github.com/6022-labs/agentic-agent-node/src/common/helpers"
	"github.com/6022-labs/agentic-agent-node/tests/agent_reasoning_self_tools_mocks/tools_mocks/interfaces_mocks"
	"github.com/6022-labs/agentic-agent-node/tests/common_tests"
	"github.com/samber/lo"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

var (
	listToolCallerId = strings.ToLower(common_tests.RandomEthAddress().Hex())
)

type WhenExecutingListConversationsToolTestingSuite struct {
	sut *tools.ListConversationsTool

	mockListConversationsExecutor *interfaces_mocks.MockListConversationsExecutorInterface
}

func WhenExecutingListConversationsToolBeforeEach(t *testing.T) *WhenExecutingListConversationsToolTestingSuite {
	mockController := gomock.NewController(t)

	mockListConversationsExecutor := interfaces_mocks.NewMockListConversationsExecutorInterface(mockController)

	return &WhenExecutingListConversationsToolTestingSuite{
		sut:                           tools.NewListConversationsTool(mockListConversationsExecutor),
		mockListConversationsExecutor: mockListConversationsExecutor,
	}
}

func TestWhenExecutingListConversationsTool(t *testing.T) {
	t.Parallel()

	ctx := helpers.ContextWithCallerId(context.Background(), listToolCallerId)

	t.Run("Given the turn carries a caller id", func(t *testing.T) {
		t.Parallel()

		t.Run("Should hand the caller's context to the use case", func(t *testing.T) {
			t.Parallel()

			suite := WhenExecutingListConversationsToolBeforeEach(t)
			suite.mockListConversationsExecutor.EXPECT().
				Execute(ctx, gomock.Any()).
				DoAndReturn(func(
					callCtx context.Context,
					_ *knowledge_requests.ListConversationsRequest,
				) (*knowledge_responses.ListConversationsResponse, error) {
					carried, _ := helpers.CallerIdFromContext(callCtx)
					assert.Equal(t, listToolCallerId, carried)
					return &knowledge_responses.ListConversationsResponse{}, nil
				}).
				Times(1)

			_, err := suite.sut.Execute(ctx, nil)

			assert.NoError(t, err)
		})
	})

	t.Run("Given the use case fails", func(t *testing.T) {
		t.Parallel()

		t.Run("Should propagate the error", func(t *testing.T) {
			t.Parallel()

			suite := WhenExecutingListConversationsToolBeforeEach(t)
			suite.mockListConversationsExecutor.EXPECT().
				Execute(ctx, gomock.Any()).
				Return(nil, errors.New("db blew"))

			_, err := suite.sut.Execute(ctx, nil)

			assert.ErrorContains(t, err, "db blew")
		})
	})

	t.Run("Given no page arguments", func(t *testing.T) {
		t.Parallel()

		t.Run("Should expose the conversation id", func(t *testing.T) {
			t.Parallel()

			suite := WhenExecutingListConversationsToolBeforeEach(t)
			suite.mockListConversationsExecutor.EXPECT().
				Execute(ctx, gomock.Any()).
				Return(&knowledge_responses.ListConversationsResponse{
					Page:          1,
					TotalElements: 1,
					Conversations: []knowledge_responses.ConversationResponse{
						{ConversationId: "c1", Name: lo.ToPtr("ERC20 mint audit"), MessageCount: 2},
					},
				}, nil)

			out, err := suite.sut.Execute(ctx, nil)

			assert.NoError(t, err)
			conversations, ok := out["conversations"].([]any)
			assert.True(t, ok)
			assert.Len(t, conversations, 1)
			assert.Equal(t, "c1", conversations[0].(map[string]any)["conversationId"])
		})
	})

	t.Run("Given string page arguments", func(t *testing.T) {
		t.Parallel()

		t.Run("Should coerce them", func(t *testing.T) {
			t.Parallel()

			suite := WhenExecutingListConversationsToolBeforeEach(t)
			suite.mockListConversationsExecutor.EXPECT().
				Execute(ctx, gomock.Any()).
				DoAndReturn(func(
					_ context.Context,
					request *knowledge_requests.ListConversationsRequest,
				) (*knowledge_responses.ListConversationsResponse, error) {
					assert.Equal(t, uint(2), *request.Page)
					assert.Equal(t, uint(50), *request.ElementsPerPage)
					return &knowledge_responses.ListConversationsResponse{}, nil
				})

			_, err := suite.sut.Execute(ctx, map[string]any{"page": "2", "elementsPerPage": float64(50)})

			assert.NoError(t, err)
		})
	})

	t.Run("Given non-positive page arguments", func(t *testing.T) {
		t.Parallel()

		t.Run("Should leave them unset", func(t *testing.T) {
			t.Parallel()

			suite := WhenExecutingListConversationsToolBeforeEach(t)
			suite.mockListConversationsExecutor.EXPECT().
				Execute(ctx, gomock.Any()).
				DoAndReturn(func(
					_ context.Context,
					request *knowledge_requests.ListConversationsRequest,
				) (*knowledge_responses.ListConversationsResponse, error) {
					assert.Nil(t, request.Page)
					assert.Nil(t, request.ElementsPerPage)
					return &knowledge_responses.ListConversationsResponse{}, nil
				})

			_, err := suite.sut.Execute(ctx, map[string]any{"page": float64(0), "elementsPerPage": float64(-5)})

			assert.NoError(t, err)
		})
	})
}
