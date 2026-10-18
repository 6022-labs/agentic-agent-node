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
	createToolCallerId = strings.ToLower(common_tests.RandomEthAddress().Hex())
)

type WhenExecutingCreateConversationToolTestingSuite struct {
	sut *tools.CreateConversationTool

	mockCreateConversationExecutor *interfaces_mocks.MockCreateConversationExecutorInterface
}

func WhenExecutingCreateConversationToolBeforeEach(t *testing.T) *WhenExecutingCreateConversationToolTestingSuite {
	mockController := gomock.NewController(t)

	mockCreateConversationExecutor := interfaces_mocks.NewMockCreateConversationExecutorInterface(mockController)

	return &WhenExecutingCreateConversationToolTestingSuite{
		sut:                            tools.NewCreateConversationTool(mockCreateConversationExecutor),
		mockCreateConversationExecutor: mockCreateConversationExecutor,
	}
}

func TestWhenExecutingCreateConversationTool(t *testing.T) {
	t.Parallel()

	ctx := helpers.ContextWithCallerId(context.Background(), createToolCallerId)

	t.Run("Given the name argument is missing", func(t *testing.T) {
		t.Parallel()

		t.Run("Should return an argument error without calling the use case", func(t *testing.T) {
			t.Parallel()

			suite := WhenExecutingCreateConversationToolBeforeEach(t)

			_, err := suite.sut.Execute(ctx, nil)

			assert.ErrorContains(t, err, "name")
		})
	})

	t.Run("Given a caller id", func(t *testing.T) {
		t.Parallel()

		t.Run("Should hand the caller's context to the use case", func(t *testing.T) {
			t.Parallel()

			suite := WhenExecutingCreateConversationToolBeforeEach(t)
			suite.mockCreateConversationExecutor.EXPECT().
				Execute(ctx, gomock.Any()).
				DoAndReturn(func(
					callCtx context.Context,
					_ *knowledge_requests.CreateConversationRequest,
				) (*knowledge_responses.ConversationResponse, error) {
					carried, _ := helpers.CallerIdFromContext(callCtx)
					assert.Equal(t, createToolCallerId, carried)
					return &knowledge_responses.ConversationResponse{ConversationId: "c1"}, nil
				}).
				Times(1)

			_, err := suite.sut.Execute(ctx, map[string]any{"name": "Audit"})

			assert.NoError(t, err)
		})
	})

	t.Run("Given the use case fails", func(t *testing.T) {
		t.Parallel()

		t.Run("Should propagate the error", func(t *testing.T) {
			t.Parallel()

			suite := WhenExecutingCreateConversationToolBeforeEach(t)
			suite.mockCreateConversationExecutor.EXPECT().
				Execute(ctx, gomock.Any()).
				Return(nil, errors.New("name taken"))

			_, err := suite.sut.Execute(ctx, map[string]any{"name": "dup"})

			assert.ErrorContains(t, err, "name taken")
		})
	})

	t.Run("Given a name", func(t *testing.T) {
		t.Parallel()

		t.Run("Should return the generated conversation id", func(t *testing.T) {
			t.Parallel()

			suite := WhenExecutingCreateConversationToolBeforeEach(t)
			suite.mockCreateConversationExecutor.EXPECT().
				Execute(ctx, gomock.Any()).
				DoAndReturn(func(
					_ context.Context,
					request *knowledge_requests.CreateConversationRequest,
				) (*knowledge_responses.ConversationResponse, error) {
					assert.Equal(t, "ERC20 mint audit", request.Name)
					return &knowledge_responses.ConversationResponse{
						ConversationId: "c1",
						Name:           lo.ToPtr("ERC20 mint audit"),
					}, nil
				})

			out, err := suite.sut.Execute(ctx, map[string]any{"name": "ERC20 mint audit"})

			assert.NoError(t, err)
			assert.Equal(t, "c1", out["conversationId"])
		})
	})
}
