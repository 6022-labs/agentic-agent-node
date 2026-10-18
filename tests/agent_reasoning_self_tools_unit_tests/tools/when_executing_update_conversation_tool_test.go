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
	updateToolCallerId = strings.ToLower(common_tests.RandomEthAddress().Hex())
)

type WhenExecutingUpdateConversationToolTestingSuite struct {
	sut *tools.UpdateConversationTool

	mockUpdateConversationExecutor *interfaces_mocks.MockUpdateConversationExecutorInterface
}

func WhenExecutingUpdateConversationToolBeforeEach(t *testing.T) *WhenExecutingUpdateConversationToolTestingSuite {
	mockController := gomock.NewController(t)

	mockUpdateConversationExecutor := interfaces_mocks.NewMockUpdateConversationExecutorInterface(mockController)

	return &WhenExecutingUpdateConversationToolTestingSuite{
		sut:                            tools.NewUpdateConversationTool(mockUpdateConversationExecutor),
		mockUpdateConversationExecutor: mockUpdateConversationExecutor,
	}
}

func TestWhenExecutingUpdateConversationTool(t *testing.T) {
	t.Parallel()

	ctx := helpers.ContextWithCallerId(context.Background(), updateToolCallerId)

	t.Run("Given the conversationId argument is missing", func(t *testing.T) {
		t.Parallel()

		t.Run("Should return an argument error without calling the use case", func(t *testing.T) {
			t.Parallel()

			suite := WhenExecutingUpdateConversationToolBeforeEach(t)

			_, err := suite.sut.Execute(ctx, map[string]any{"name": "Solidity audits"})

			assert.ErrorContains(t, err, "conversationId")
		})
	})

	t.Run("Given the name argument is missing", func(t *testing.T) {
		t.Parallel()

		t.Run("Should return an argument error without calling the use case", func(t *testing.T) {
			t.Parallel()

			suite := WhenExecutingUpdateConversationToolBeforeEach(t)

			_, err := suite.sut.Execute(ctx, map[string]any{"conversationId": "c1"})

			assert.ErrorContains(t, err, "name")
		})
	})

	t.Run("Given a caller id", func(t *testing.T) {
		t.Parallel()

		t.Run("Should hand the caller's context to the use case", func(t *testing.T) {
			t.Parallel()

			suite := WhenExecutingUpdateConversationToolBeforeEach(t)
			suite.mockUpdateConversationExecutor.EXPECT().
				Execute(ctx, gomock.Any()).
				DoAndReturn(func(
					callCtx context.Context,
					_ *knowledge_requests.UpdateConversationRequest,
				) (*knowledge_responses.ConversationResponse, error) {
					carried, _ := helpers.CallerIdFromContext(callCtx)
					assert.Equal(t, updateToolCallerId, carried)
					return &knowledge_responses.ConversationResponse{ConversationId: "c1"}, nil
				}).
				Times(1)

			_, err := suite.sut.Execute(ctx, map[string]any{"conversationId": "c1", "name": "Audit"})

			assert.NoError(t, err)
		})
	})

	t.Run("Given the use case fails", func(t *testing.T) {
		t.Parallel()

		t.Run("Should propagate the error", func(t *testing.T) {
			t.Parallel()

			suite := WhenExecutingUpdateConversationToolBeforeEach(t)
			suite.mockUpdateConversationExecutor.EXPECT().
				Execute(ctx, gomock.Any()).
				Return(nil, errors.New("not found"))

			_, err := suite.sut.Execute(ctx, map[string]any{"conversationId": "c1", "name": "x"})

			assert.ErrorContains(t, err, "not found")
		})
	})

	t.Run("Given both arguments", func(t *testing.T) {
		t.Parallel()

		t.Run("Should rename the session", func(t *testing.T) {
			t.Parallel()

			suite := WhenExecutingUpdateConversationToolBeforeEach(t)
			suite.mockUpdateConversationExecutor.EXPECT().
				Execute(ctx, gomock.Any()).
				DoAndReturn(func(
					_ context.Context,
					request *knowledge_requests.UpdateConversationRequest,
				) (*knowledge_responses.ConversationResponse, error) {
					assert.Equal(t, "c1", request.ConversationId)
					assert.Equal(t, "Solidity audits", request.Name)
					return &knowledge_responses.ConversationResponse{
						ConversationId: "c1",
						Name:           lo.ToPtr("Solidity audits"),
					}, nil
				})

			out, err := suite.sut.Execute(ctx, map[string]any{
				"conversationId": "c1",
				"name":           "Solidity audits",
			})

			assert.NoError(t, err)
			assert.Equal(t, "Solidity audits", out["name"])
		})
	})
}
