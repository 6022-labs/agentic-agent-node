package tools_test

import (
	"context"
	"errors"
	"testing"

	reasoning_requests "github.com/6022-labs/agentic-agent-node/src/agent_reasoning/use_cases/requests"
	reasoning_responses "github.com/6022-labs/agentic-agent-node/src/agent_reasoning/use_cases/responses"
	"github.com/6022-labs/agentic-agent-node/src/agent_reasoning_self_tools/tools"
	"github.com/6022-labs/agentic-agent-node/tests/agent_reasoning_self_tools_mocks/tools_mocks/interfaces_mocks"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

type WhenExecutingSetAgentGuidanceToolTestingSuite struct {
	sut *tools.SetAgentGuidanceTool

	mockSetAgentGuidanceExecutor *interfaces_mocks.MockSetAgentGuidanceExecutorInterface
}

func WhenExecutingSetAgentGuidanceToolBeforeEach(t *testing.T) *WhenExecutingSetAgentGuidanceToolTestingSuite {
	mockController := gomock.NewController(t)

	mockSetAgentGuidanceExecutor := interfaces_mocks.NewMockSetAgentGuidanceExecutorInterface(mockController)

	return &WhenExecutingSetAgentGuidanceToolTestingSuite{
		sut:                          tools.NewSetAgentGuidanceTool(mockSetAgentGuidanceExecutor),
		mockSetAgentGuidanceExecutor: mockSetAgentGuidanceExecutor,
	}
}

func TestWhenExecutingSetAgentGuidanceTool(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	t.Run("Given the content argument is missing", func(t *testing.T) {
		t.Parallel()

		t.Run("Should fail without calling the use case", func(t *testing.T) {
			t.Parallel()

			suite := WhenExecutingSetAgentGuidanceToolBeforeEach(t)

			out, err := suite.sut.Execute(ctx, nil)

			assert.Nil(t, out)
			assert.ErrorContains(t, err, "content")
		})
	})

	t.Run("Given the use case fails", func(t *testing.T) {
		t.Parallel()

		t.Run("Should propagate the error", func(t *testing.T) {
			t.Parallel()

			suite := WhenExecutingSetAgentGuidanceToolBeforeEach(t)
			suite.mockSetAgentGuidanceExecutor.EXPECT().
				Execute(ctx, gomock.Any()).
				Return(nil, errors.New("db down"))

			out, err := suite.sut.Execute(ctx, map[string]any{"content": "x"})

			assert.Nil(t, out)
			assert.ErrorContains(t, err, "db down")
		})
	})

	t.Run("Given a content argument", func(t *testing.T) {
		t.Parallel()

		t.Run("Should forward it to the use case", func(t *testing.T) {
			t.Parallel()

			suite := WhenExecutingSetAgentGuidanceToolBeforeEach(t)
			suite.mockSetAgentGuidanceExecutor.EXPECT().
				Execute(ctx, &reasoning_requests.SetAgentGuidanceRequest{Content: "answer concisely"}).
				Return(&reasoning_responses.AgentGuidanceResponse{Content: "answer concisely"}, nil)

			out, err := suite.sut.Execute(ctx, map[string]any{"content": "answer concisely"})

			assert.NoError(t, err)
			if assert.NotNil(t, out) {
				assert.Equal(t, "answer concisely", out["content"])
			}
		})
	})
}
