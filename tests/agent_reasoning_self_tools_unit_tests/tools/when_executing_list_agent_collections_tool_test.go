package tools_test

import (
	"context"
	"errors"
	"testing"

	identity_requests "github.com/6022-labs/agentic-agent-node/src/agent_identity/use_cases/requests"
	identity_responses "github.com/6022-labs/agentic-agent-node/src/agent_identity/use_cases/responses"
	"github.com/6022-labs/agentic-agent-node/src/agent_reasoning_self_tools/tools"
	"github.com/6022-labs/agentic-agent-node/tests/agent_reasoning_self_tools_mocks/tools_mocks/interfaces_mocks"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

type WhenExecutingListAgentCollectionsToolTestingSuite struct {
	sut *tools.ListAgentCollectionsTool

	mockListAgentCollectionsExecutor *interfaces_mocks.MockListAgentCollectionsExecutorInterface
}

func WhenExecutingListAgentCollectionsToolBeforeEach(t *testing.T) *WhenExecutingListAgentCollectionsToolTestingSuite {
	mockController := gomock.NewController(t)

	mockListAgentCollectionsExecutor := interfaces_mocks.NewMockListAgentCollectionsExecutorInterface(mockController)

	return &WhenExecutingListAgentCollectionsToolTestingSuite{
		sut:                              tools.NewListAgentCollectionsTool(mockListAgentCollectionsExecutor),
		mockListAgentCollectionsExecutor: mockListAgentCollectionsExecutor,
	}
}

func TestWhenExecutingListAgentCollectionsTool(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	t.Run("Given the chainId argument is missing", func(t *testing.T) {
		t.Parallel()

		t.Run("Should fail without calling the use case", func(t *testing.T) {
			t.Parallel()

			suite := WhenExecutingListAgentCollectionsToolBeforeEach(t)

			out, err := suite.sut.Execute(ctx, map[string]any{})

			assert.Nil(t, out)
			assert.ErrorContains(t, err, "chainId")
		})
	})

	t.Run("Given the use case fails", func(t *testing.T) {
		t.Parallel()

		t.Run("Should propagate the error", func(t *testing.T) {
			t.Parallel()

			suite := WhenExecutingListAgentCollectionsToolBeforeEach(t)
			suite.mockListAgentCollectionsExecutor.EXPECT().
				Execute(ctx, gomock.Any()).
				Return(nil, errors.New("chain unreachable"))

			out, err := suite.sut.Execute(ctx, map[string]any{"chainId": "80002"})

			assert.Nil(t, out)
			assert.ErrorContains(t, err, "chain unreachable")
		})
	})

	t.Run("Given a chainId argument", func(t *testing.T) {
		t.Parallel()

		t.Run("Should forward it to the use case", func(t *testing.T) {
			t.Parallel()

			suite := WhenExecutingListAgentCollectionsToolBeforeEach(t)
			suite.mockListAgentCollectionsExecutor.EXPECT().
				Execute(ctx, &identity_requests.ListAgentCollectionsRequest{
					ChainIdRequest: identity_requests.ChainIdRequest{ChainId: "80002"},
				}).
				Return(&identity_responses.AgentCollectionsResponse{}, nil)

			out, err := suite.sut.Execute(ctx, map[string]any{"chainId": "80002"})

			assert.NoError(t, err)
			assert.NotNil(t, out)
		})
	})
}
