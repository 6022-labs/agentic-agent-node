package use_cases_test

import (
	"context"
	"errors"
	"testing"

	knowledge_interfaces "github.com/6022-labs/agentic-agent-node/src/agent_knowledge/services/interfaces"
	traces_interfaces "github.com/6022-labs/agentic-agent-node/src/agent_reasoning/traces/interfaces"
	"github.com/6022-labs/agentic-agent-node/src/agent_reasoning/use_cases"
	"github.com/6022-labs/agentic-agent-node/src/agent_reasoning/use_cases/requests"
	apperrors "github.com/6022-labs/agentic-agent-node/src/common/errors"
	metrics_interfaces_mocks "github.com/6022-labs/agentic-agent-node/tests/agent_reasoning_mocks/metrics_mocks/interfaces_mocks"
	"github.com/6022-labs/agentic-agent-node/tests/agent_reasoning_mocks/services_mocks/interfaces_mocks"
	traces_interfaces_mocks "github.com/6022-labs/agentic-agent-node/tests/agent_reasoning_mocks/traces_mocks/interfaces_mocks"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

type WhenCreatingResponseTestingSuite struct {
	sut *use_cases.CreateResponse

	mockResponseDriver *interfaces_mocks.MockResponseDriverInterface
}

func WhenCreatingResponseBeforeEach(t *testing.T) *WhenCreatingResponseTestingSuite {
	mockController := gomock.NewController(t)

	mockResponseDriver := interfaces_mocks.NewMockResponseDriverInterface(mockController)

	mockResponseMetrics := metrics_interfaces_mocks.NewMockResponseMetricsInterface(mockController)
	mockResponseMetrics.EXPECT().RecordResponseSucceeded(gomock.Any()).AnyTimes()
	mockResponseMetrics.EXPECT().RecordResponseFailed(gomock.Any()).AnyTimes()

	mockSpan := traces_interfaces_mocks.NewMockSpan(mockController)
	mockSpan.EXPECT().Fail(gomock.Any()).AnyTimes()
	mockSpan.EXPECT().End().AnyTimes()

	mockReasoningTracer := traces_interfaces_mocks.NewMockReasoningTracerInterface(mockController)
	mockReasoningTracer.EXPECT().StartInference(gomock.Any(), gomock.Any(), gomock.Any()).
		DoAndReturn(func(ctx context.Context, _, _ string) (context.Context, traces_interfaces.Span) {
			return ctx, mockSpan
		}).AnyTimes()

	sut := use_cases.NewCreateResponse(mockResponseDriver, mockResponseMetrics, mockReasoningTracer)

	return &WhenCreatingResponseTestingSuite{
		sut:                sut,
		mockResponseDriver: mockResponseDriver,
	}
}

func userInputRequest(conversationId, content string) *requests.CreateResponseRequest {
	return &requests.CreateResponseRequest{
		PreviousResponseId: conversationId,
		Input:              []requests.InputMessage{{Role: "user", Content: content}},
	}
}

func TestWhenCreatingResponse(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	conversationId := "conv-1"

	t.Run("Given the request is invalid", func(t *testing.T) {
		t.Parallel()

		t.Run("Should return a validation error without touching dependencies", func(t *testing.T) {
			t.Parallel()

			suite := WhenCreatingResponseBeforeEach(t)

			_, err := suite.sut.Execute(ctx, &requests.CreateResponseRequest{})

			var validationErr *apperrors.ValidationError
			assert.ErrorAs(t, err, &validationErr)
		})
	})

	t.Run("Given the request carries instructions and the driver persists an assistant message", func(t *testing.T) {
		t.Parallel()

		expectedFinal := "hi there"

		initSuite := func(suite *WhenCreatingResponseTestingSuite) {
			suite.mockResponseDriver.EXPECT().
				Drive(gomock.Any(), conversationId, "be nice", gomock.Any()).
				DoAndReturn(func(_ context.Context, _ string, _ string, inputs []knowledge_interfaces.ConversationMessage) (*knowledge_interfaces.ConversationMessage, error) {
					if assert.Len(t, inputs, 1) {
						assert.Equal(t, "user", inputs[0].Role)
						assert.Equal(t, "hello", inputs[0].Content)
					}
					return &knowledge_interfaces.ConversationMessage{Role: "assistant", Content: expectedFinal}, nil
				})
		}

		t.Run(
			"Should forward instructions as the overlay and surface the assistant message as an OpenAI output item",
			func(t *testing.T) {
				t.Parallel()

				suite := WhenCreatingResponseBeforeEach(t)
				initSuite(suite)

				resp, err := suite.sut.Execute(ctx, &requests.CreateResponseRequest{
					PreviousResponseId: conversationId,
					Instructions:       "be nice",
					Input:              []requests.InputMessage{{Role: "user", Content: "hello"}},
				})

				assert.NoError(t, err)
				assert.Equal(t, "response", resp.Object)
				assert.Equal(t, "completed", resp.Status)
				assert.Equal(t, conversationId, resp.Id)
				if assert.Len(t, resp.Output, 1) {
					assert.Equal(t, "message", resp.Output[0].Type)
					if assert.Len(t, resp.Output[0].Content, 1) {
						assert.Equal(t, "output_text", resp.Output[0].Content[0].Type)
						assert.Equal(t, expectedFinal, resp.Output[0].Content[0].Text)
					}
				}
			},
		)
	})

	t.Run("Given no conversation key is provided", func(t *testing.T) {
		t.Parallel()

		initSuite := func(suite *WhenCreatingResponseTestingSuite) {
			suite.mockResponseDriver.EXPECT().
				Drive(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
				DoAndReturn(func(_ context.Context, conversationId string, _ string, _ []knowledge_interfaces.ConversationMessage) (*knowledge_interfaces.ConversationMessage, error) {
					assert.NotEmpty(t, conversationId)
					return &knowledge_interfaces.ConversationMessage{Role: "assistant", Content: "hi"}, nil
				})
		}

		t.Run("Should mint a conversation id and echo it as the response id", func(t *testing.T) {
			t.Parallel()

			suite := WhenCreatingResponseBeforeEach(t)
			initSuite(suite)

			resp, err := suite.sut.Execute(ctx, &requests.CreateResponseRequest{
				Input: []requests.InputMessage{{Role: "user", Content: "hi"}},
			})

			assert.NoError(t, err)
			assert.NotEmpty(t, resp.Id)
			assert.Nil(t, resp.PreviousResponseId)
		})
	})

	t.Run("Given the response driver returns nil", func(t *testing.T) {
		t.Parallel()

		initSuite := func(suite *WhenCreatingResponseTestingSuite) {
			suite.mockResponseDriver.EXPECT().
				Drive(gomock.Any(), conversationId, gomock.Any(), gomock.Any()).
				Return(nil, nil)
		}

		t.Run("Should return status=incomplete with no output", func(t *testing.T) {
			t.Parallel()

			suite := WhenCreatingResponseBeforeEach(t)
			initSuite(suite)

			resp, err := suite.sut.Execute(ctx, userInputRequest(conversationId, "hello"))

			assert.NoError(t, err)
			assert.Equal(t, "incomplete", resp.Status)
			assert.Empty(t, resp.Output)
		})
	})

	t.Run("Given the response driver fails", func(t *testing.T) {
		t.Parallel()

		initSuite := func(suite *WhenCreatingResponseTestingSuite) {
			suite.mockResponseDriver.EXPECT().
				Drive(gomock.Any(), conversationId, gomock.Any(), gomock.Any()).
				Return(nil, errors.New("driver broke"))
		}

		t.Run("Should return InternalError carrying the driver message", func(t *testing.T) {
			t.Parallel()

			suite := WhenCreatingResponseBeforeEach(t)
			initSuite(suite)

			_, err := suite.sut.Execute(ctx, userInputRequest(conversationId, "hello"))

			var internalErr *apperrors.InternalError
			assert.ErrorAs(t, err, &internalErr)
			assert.Contains(t, internalErr.Message, "driver broke")
		})
	})
}
