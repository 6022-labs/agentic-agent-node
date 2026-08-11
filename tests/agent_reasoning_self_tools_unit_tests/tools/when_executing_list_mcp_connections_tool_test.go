package tools_test

import (
	"context"
	"errors"
	"testing"

	reasoning_responses "github.com/6022-labs/agentic-agent-node/src/agent_reasoning/use_cases/responses"
	"github.com/6022-labs/agentic-agent-node/src/agent_reasoning_self_tools/tools"
	"github.com/6022-labs/agentic-agent-node/tests/agent_reasoning_self_tools_mocks/tools_mocks/interfaces_mocks"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

type WhenExecutingListMCPConnectionsToolTestingSuite struct {
	sut *tools.ListMCPConnectionsTool

	mockListMCPConnectionsExecutor *interfaces_mocks.MockListMCPConnectionsExecutorInterface
}

func WhenExecutingListMCPConnectionsToolBeforeEach(t *testing.T) *WhenExecutingListMCPConnectionsToolTestingSuite {
	mockController := gomock.NewController(t)

	mockListMCPConnectionsExecutor := interfaces_mocks.NewMockListMCPConnectionsExecutorInterface(mockController)

	return &WhenExecutingListMCPConnectionsToolTestingSuite{
		sut:                            tools.NewListMCPConnectionsTool(mockListMCPConnectionsExecutor),
		mockListMCPConnectionsExecutor: mockListMCPConnectionsExecutor,
	}
}

func TestWhenExecutingListMCPConnectionsTool(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	t.Run("Given the use case succeeds", func(t *testing.T) {
		t.Parallel()

		t.Run("Should return the encoded response", func(t *testing.T) {
			t.Parallel()

			suite := WhenExecutingListMCPConnectionsToolBeforeEach(t)
			suite.mockListMCPConnectionsExecutor.EXPECT().
				Execute(ctx).
				Return(&reasoning_responses.MCPConnectionListResponse{}, nil)

			out, err := suite.sut.Execute(ctx, nil)

			assert.NoError(t, err)
			assert.NotNil(t, out)
		})
	})

	t.Run("Given the use case fails", func(t *testing.T) {
		t.Parallel()

		t.Run("Should propagate the error", func(t *testing.T) {
			t.Parallel()

			suite := WhenExecutingListMCPConnectionsToolBeforeEach(t)
			suite.mockListMCPConnectionsExecutor.EXPECT().Execute(ctx).Return(nil, errors.New("boom"))

			out, err := suite.sut.Execute(ctx, nil)

			assert.Nil(t, out)
			assert.ErrorContains(t, err, "boom")
		})
	})
}
