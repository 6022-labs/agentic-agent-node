package tools_test

import (
	"context"
	"testing"

	reasoning_requests "github.com/6022-labs/agentic-agent-node/src/agent_reasoning/use_cases/requests"
	reasoning_responses "github.com/6022-labs/agentic-agent-node/src/agent_reasoning/use_cases/responses"
	"github.com/6022-labs/agentic-agent-node/src/agent_reasoning_self_tools/tools"
	apperrors "github.com/6022-labs/agentic-agent-node/src/common/errors"
	"github.com/6022-labs/agentic-agent-node/tests/agent_reasoning_self_tools_mocks/tools_mocks/interfaces_mocks"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

type WhenExecutingCallAgentToolTestingSuite struct {
	sut *tools.CallAgentTool

	mockCallAgentExecutor *interfaces_mocks.MockCallAgentExecutorInterface
}

func WhenExecutingCallAgentToolBeforeEach(t *testing.T) *WhenExecutingCallAgentToolTestingSuite {
	mockController := gomock.NewController(t)

	mockCallAgentExecutor := interfaces_mocks.NewMockCallAgentExecutorInterface(mockController)

	return &WhenExecutingCallAgentToolTestingSuite{
		sut:                   tools.NewCallAgentTool(mockCallAgentExecutor),
		mockCallAgentExecutor: mockCallAgentExecutor,
	}
}

func TestWhenExecutingCallAgentTool(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	t.Run("Given the url argument is missing", func(t *testing.T) {
		t.Parallel()

		args := map[string]any{"message": "hello"}

		t.Run("Should fail without calling the use case", func(t *testing.T) {
			t.Parallel()

			suite := WhenExecutingCallAgentToolBeforeEach(t)

			out, err := suite.sut.Execute(ctx, args)

			assert.Nil(t, out)
			assert.ErrorContains(t, err, "url")
		})
	})

	t.Run("Given the message argument is missing", func(t *testing.T) {
		t.Parallel()

		args := map[string]any{"url": "https://peer.example"}

		t.Run("Should fail without calling the use case", func(t *testing.T) {
			t.Parallel()

			suite := WhenExecutingCallAgentToolBeforeEach(t)

			out, err := suite.sut.Execute(ctx, args)

			assert.Nil(t, out)
			assert.ErrorContains(t, err, "message")
		})
	})

	t.Run("Given the use case fails", func(t *testing.T) {
		t.Parallel()

		args := map[string]any{"url": "https://peer.example", "message": "hello"}

		initSuite := func(suite *WhenExecutingCallAgentToolTestingSuite) {
			suite.mockCallAgentExecutor.EXPECT().
				Execute(ctx, gomock.Any()).
				Return(nil, apperrors.NewUnavailableError("agent_call_failed", "the agent could not be reached", nil)).
				Times(1)
		}

		t.Run("Should propagate the error", func(t *testing.T) {
			t.Parallel()

			suite := WhenExecutingCallAgentToolBeforeEach(t)
			initSuite(suite)

			out, err := suite.sut.Execute(ctx, args)

			assert.Nil(t, out)
			var unavailableError *apperrors.UnavailableError
			assert.ErrorAs(t, err, &unavailableError)
			assert.Equal(t, "agent_call_failed", unavailableError.Code)
		})
	})

	t.Run("Given every argument", func(t *testing.T) {
		t.Parallel()

		args := map[string]any{
			"url":              "https://peer.example",
			"message":          "hello",
			"contextId":        "conversation-1",
			"paymentSignature": "eyJ4NDAy",
		}

		initSuite := func(suite *WhenExecutingCallAgentToolTestingSuite) {
			suite.mockCallAgentExecutor.EXPECT().
				Execute(ctx, &reasoning_requests.CallAgentRequest{
					Url:              "https://peer.example",
					Message:          "hello",
					ContextId:        "conversation-1",
					PaymentSignature: "eyJ4NDAy",
				}).
				Return(&reasoning_responses.CallAgentResponse{
					Status: reasoning_responses.CallAgentStatusOk,
					Text:   "hi",
				}, nil).
				Times(1)
		}

		t.Run("Should return the reply", func(t *testing.T) {
			t.Parallel()

			suite := WhenExecutingCallAgentToolBeforeEach(t)
			initSuite(suite)

			out, err := suite.sut.Execute(ctx, args)

			assert.NoError(t, err)
			assert.Equal(t, "hi", out["text"])
			assert.Equal(t, reasoning_responses.CallAgentStatusOk, out["status"])
		})
	})
}
