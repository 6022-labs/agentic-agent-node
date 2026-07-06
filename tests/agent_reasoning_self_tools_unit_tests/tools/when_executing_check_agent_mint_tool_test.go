package tools_test

import (
	"context"
	"errors"
	"testing"

	identity_responses "github.com/6022-labs/agentic-agent-node/src/agent_identity/use_cases/responses"
	"github.com/6022-labs/agentic-agent-node/src/agent_reasoning_self_tools/tools"
	"github.com/6022-labs/agentic-agent-node/tests/agent_reasoning_self_tools_mocks/tools_mocks/interfaces_mocks"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

type WhenExecutingCheckAgentMintToolTestingSuite struct {
	sut *tools.CheckAgentMintTool

	mockCheckAgentMintExecutor *interfaces_mocks.MockCheckAgentMintExecutorInterface
}

func WhenExecutingCheckAgentMintToolBeforeEach(t *testing.T) *WhenExecutingCheckAgentMintToolTestingSuite {
	mockController := gomock.NewController(t)

	mockCheckAgentMintExecutor := interfaces_mocks.NewMockCheckAgentMintExecutorInterface(mockController)

	return &WhenExecutingCheckAgentMintToolTestingSuite{
		sut:                        tools.NewCheckAgentMintTool(mockCheckAgentMintExecutor),
		mockCheckAgentMintExecutor: mockCheckAgentMintExecutor,
	}
}

func TestWhenExecutingCheckAgentMintTool(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	t.Run("Given the use case succeeds", func(t *testing.T) {
		t.Parallel()

		t.Run("Should return the encoded response", func(t *testing.T) {
			t.Parallel()

			suite := WhenExecutingCheckAgentMintToolBeforeEach(t)
			suite.mockCheckAgentMintExecutor.EXPECT().Execute(ctx).Return(&identity_responses.RegistrationStatusResponse{}, nil)

			out, err := suite.sut.Execute(ctx, nil)

			assert.NoError(t, err)
			assert.NotNil(t, out)
		})
	})

	t.Run("Given the use case fails", func(t *testing.T) {
		t.Parallel()

		t.Run("Should propagate the error", func(t *testing.T) {
			t.Parallel()

			suite := WhenExecutingCheckAgentMintToolBeforeEach(t)
			suite.mockCheckAgentMintExecutor.EXPECT().Execute(ctx).Return(nil, errors.New("boom"))

			out, err := suite.sut.Execute(ctx, nil)

			assert.Nil(t, out)
			assert.ErrorContains(t, err, "boom")
		})
	})
}
