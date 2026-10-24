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

type WhenExecutingGetLLMConnectionToolTestingSuite struct {
	sut *tools.GetLLMConnectionTool

	mockGetLLMConnectionExecutor *interfaces_mocks.MockGetLLMConnectionExecutorInterface
}

func WhenExecutingGetLLMConnectionToolBeforeEach(t *testing.T) *WhenExecutingGetLLMConnectionToolTestingSuite {
	mockController := gomock.NewController(t)

	mockGetLLMConnectionExecutor := interfaces_mocks.NewMockGetLLMConnectionExecutorInterface(mockController)

	return &WhenExecutingGetLLMConnectionToolTestingSuite{
		sut:                          tools.NewGetLLMConnectionTool(mockGetLLMConnectionExecutor),
		mockGetLLMConnectionExecutor: mockGetLLMConnectionExecutor,
	}
}

func TestWhenExecutingGetLLMConnectionTool(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	t.Run("Given the use case fails", func(t *testing.T) {
		t.Parallel()

		t.Run("Should propagate the error", func(t *testing.T) {
			t.Parallel()

			suite := WhenExecutingGetLLMConnectionToolBeforeEach(t)
			suite.mockGetLLMConnectionExecutor.EXPECT().Execute(ctx).Return(nil, errors.New("boom"))

			out, err := suite.sut.Execute(ctx, nil)

			assert.Nil(t, out)
			assert.ErrorContains(t, err, "boom")
		})
	})

	t.Run("Given the use case succeeds", func(t *testing.T) {
		t.Parallel()

		t.Run("Should return the encoded response", func(t *testing.T) {
			t.Parallel()

			suite := WhenExecutingGetLLMConnectionToolBeforeEach(t)
			suite.mockGetLLMConnectionExecutor.EXPECT().
				Execute(ctx).
				Return(&reasoning_responses.LLMConnectionResponse{}, nil)

			out, err := suite.sut.Execute(ctx, nil)

			assert.NoError(t, err)
			assert.NotNil(t, out)
		})
	})
}
