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

type WhenExecutingGetAgentRegistrationToolTestingSuite struct {
	sut *tools.GetAgentRegistrationTool

	mockGetAgentRegistrationExecutor *interfaces_mocks.MockGetAgentRegistrationExecutorInterface
}

func WhenExecutingGetAgentRegistrationToolBeforeEach(t *testing.T) *WhenExecutingGetAgentRegistrationToolTestingSuite {
	mockController := gomock.NewController(t)

	mockGetAgentRegistrationExecutor := interfaces_mocks.NewMockGetAgentRegistrationExecutorInterface(mockController)

	return &WhenExecutingGetAgentRegistrationToolTestingSuite{
		sut:                              tools.NewGetAgentRegistrationTool(mockGetAgentRegistrationExecutor),
		mockGetAgentRegistrationExecutor: mockGetAgentRegistrationExecutor,
	}
}

func TestWhenExecutingGetAgentRegistrationTool(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	t.Run("Given the use case succeeds", func(t *testing.T) {
		t.Parallel()

		t.Run("Should return the encoded response", func(t *testing.T) {
			t.Parallel()

			suite := WhenExecutingGetAgentRegistrationToolBeforeEach(t)
			suite.mockGetAgentRegistrationExecutor.EXPECT().Execute(ctx).Return(&identity_responses.AgentRegistrationResponse{}, nil)

			out, err := suite.sut.Execute(ctx, nil)

			assert.NoError(t, err)
			assert.NotNil(t, out)
		})
	})

	t.Run("Given the use case fails", func(t *testing.T) {
		t.Parallel()

		t.Run("Should propagate the error", func(t *testing.T) {
			t.Parallel()

			suite := WhenExecutingGetAgentRegistrationToolBeforeEach(t)
			suite.mockGetAgentRegistrationExecutor.EXPECT().Execute(ctx).Return(nil, errors.New("boom"))

			out, err := suite.sut.Execute(ctx, nil)

			assert.Nil(t, out)
			assert.ErrorContains(t, err, "boom")
		})
	})
}
