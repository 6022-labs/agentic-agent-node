package use_cases_test

import (
	"context"
	"errors"
	"strings"
	"testing"

	knowledge_interfaces "github.com/6022-labs/agentic-agent-node/src/agent_knowledge/services/interfaces"
	traces_interfaces "github.com/6022-labs/agentic-agent-node/src/agent_reasoning/traces/interfaces"
	"github.com/6022-labs/agentic-agent-node/src/agent_reasoning/use_cases"
	"github.com/6022-labs/agentic-agent-node/src/agent_reasoning/use_cases/requests"
	apperrors "github.com/6022-labs/agentic-agent-node/src/common/errors"
	"github.com/6022-labs/agentic-agent-node/src/common/helpers"
	knowledge_interfaces_mocks "github.com/6022-labs/agentic-agent-node/tests/agent_knowledge_mocks/services_mocks/interfaces_mocks"
	metrics_interfaces_mocks "github.com/6022-labs/agentic-agent-node/tests/agent_reasoning_mocks/metrics_mocks/interfaces_mocks"
	"github.com/6022-labs/agentic-agent-node/tests/agent_reasoning_mocks/services_mocks/interfaces_mocks"
	traces_interfaces_mocks "github.com/6022-labs/agentic-agent-node/tests/agent_reasoning_mocks/traces_mocks/interfaces_mocks"
	"github.com/6022-labs/agentic-agent-node/tests/common_tests"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

var (
	responseCallerId  = strings.ToLower(common_tests.RandomEthAddress().Hex())
	responseCallerCtx = helpers.ContextWithCallerId(context.Background(), responseCallerId)
)

type WhenCreatingResponseTestingSuite struct {
	sut *use_cases.CreateResponse

	mockResponseDriver       *interfaces_mocks.MockResponseDriverInterface
	mockConversationResolver *knowledge_interfaces_mocks.MockConversationResolverInterface
	mockResponseMetrics      *metrics_interfaces_mocks.MockResponseMetricsInterface
}

func WhenCreatingResponseBeforeEach(t *testing.T) *WhenCreatingResponseTestingSuite {
	mockController := gomock.NewController(t)

	mockResponseDriver := interfaces_mocks.NewMockResponseDriverInterface(mockController)
	mockConversationResolver := knowledge_interfaces_mocks.NewMockConversationResolverInterface(mockController)

	mockResponseMetrics := metrics_interfaces_mocks.NewMockResponseMetricsInterface(mockController)
	mockResponseMetrics.EXPECT().RecordResponseSucceeded(gomock.Any()).AnyTimes()

	mockSpan := traces_interfaces_mocks.NewMockSpan(mockController)
	mockSpan.EXPECT().Fail(gomock.Any()).AnyTimes()
	mockSpan.EXPECT().End().AnyTimes()

	mockReasoningTracer := traces_interfaces_mocks.NewMockReasoningTracerInterface(mockController)
	mockReasoningTracer.EXPECT().StartInference(gomock.Any(), gomock.Any(), gomock.Any()).
		DoAndReturn(func(ctx context.Context, _, _ string) (context.Context, traces_interfaces.Span) {
			return ctx, mockSpan
		}).AnyTimes()

	sut := use_cases.NewCreateResponse(
		mockResponseDriver,
		mockConversationResolver,
		mockResponseMetrics,
		mockReasoningTracer,
	)

	return &WhenCreatingResponseTestingSuite{
		sut:                      sut,
		mockResponseDriver:       mockResponseDriver,
		mockConversationResolver: mockConversationResolver,
		mockResponseMetrics:      mockResponseMetrics,
	}
}

// expectOwnedKey resolves the requested key to itself, as for the caller's own session.
func (s *WhenCreatingResponseTestingSuite) expectOwnedKey(conversationId string) {
	s.mockConversationResolver.EXPECT().
		Resolve(gomock.Any(), conversationId, responseCallerId).
		Return(conversationId, nil)
}

func (s *WhenCreatingResponseTestingSuite) expectDriverContextCallerId(t *testing.T, conversationId string) {
	s.mockResponseDriver.EXPECT().
		Drive(gomock.Any(), conversationId, gomock.Any(), gomock.Any()).
		DoAndReturn(func(
			driveCtx context.Context,
			_ string,
			_ string,
			_ []knowledge_interfaces.ConversationMessage,
		) (*knowledge_interfaces.ConversationMessage, error) {
			callerId, ok := helpers.CallerIdFromContext(driveCtx)
			assert.True(t, ok)
			assert.Equal(t, responseCallerId, callerId)
			return nil, nil
		})
}

func userInputRequest(conversationId, content string) *requests.CreateResponseRequest {
	return &requests.CreateResponseRequest{
		PreviousResponseId: conversationId,
		Input:              []requests.InputMessage{{Role: "user", Content: content}},
	}
}

func TestWhenCreatingResponse(t *testing.T) {
	t.Parallel()

	conversationId := "conv-1"

	t.Run("Given an invalid request", func(t *testing.T) {
		t.Parallel()

		t.Run("Should return a ValidationError without touching dependencies", func(t *testing.T) {
			t.Parallel()

			suite := WhenCreatingResponseBeforeEach(t)

			_, err := suite.sut.Execute(responseCallerCtx, &requests.CreateResponseRequest{})

			var validationErr *apperrors.ValidationError
			assert.ErrorAs(t, err, &validationErr)
		})
	})

	t.Run("Given no caller id", func(t *testing.T) {
		t.Parallel()

		t.Run("Should return Unauthorized without touching dependencies", func(t *testing.T) {
			t.Parallel()

			suite := WhenCreatingResponseBeforeEach(t)

			_, err := suite.sut.Execute(context.Background(), userInputRequest(conversationId, "hello"))

			var unauthorizedErr *apperrors.UnauthorizedError
			assert.ErrorAs(t, err, &unauthorizedErr)
			assert.Equal(t, "missing_caller_id", unauthorizedErr.Code)
		})
	})

	t.Run("Given the conversation is out of the caller's view", func(t *testing.T) {
		t.Parallel()

		initSuite := func(suite *WhenCreatingResponseTestingSuite) {
			suite.mockConversationResolver.EXPECT().
				Resolve(gomock.Any(), conversationId, responseCallerId).
				Return("", knowledge_interfaces.ErrConversationUnavailable)
			suite.mockResponseMetrics.EXPECT().RecordResponseFailed(gomock.Any()).AnyTimes()
		}

		t.Run("Should return the conversation_unavailable conflict", func(t *testing.T) {
			t.Parallel()

			suite := WhenCreatingResponseBeforeEach(t)
			initSuite(suite)

			_, err := suite.sut.Execute(responseCallerCtx, userInputRequest(conversationId, "hello"))

			var conflictErr *apperrors.ConflictError
			if assert.ErrorAs(t, err, &conflictErr) {
				assert.Equal(t, "conversation_unavailable", conflictErr.Code)
			}
		})
	})

	t.Run("Given the agent is not minted yet", func(t *testing.T) {
		t.Parallel()

		initSuite := func(suite *WhenCreatingResponseTestingSuite) {
			suite.mockConversationResolver.EXPECT().
				Resolve(gomock.Any(), conversationId, responseCallerId).
				Return("", knowledge_interfaces.ErrAgentNotMinted)
			suite.mockResponseMetrics.EXPECT().RecordResponseFailed(gomock.Any()).AnyTimes()
		}

		t.Run("Should return the agent_not_minted conflict", func(t *testing.T) {
			t.Parallel()

			suite := WhenCreatingResponseBeforeEach(t)
			initSuite(suite)

			_, err := suite.sut.Execute(responseCallerCtx, userInputRequest(conversationId, "hello"))

			var conflictErr *apperrors.ConflictError
			if assert.ErrorAs(t, err, &conflictErr) {
				assert.Equal(t, "agent_not_minted", conflictErr.Code)
			}
		})
	})

	t.Run("Given the conversation resolver fails", func(t *testing.T) {
		t.Parallel()

		initSuite := func(suite *WhenCreatingResponseTestingSuite) {
			suite.mockConversationResolver.EXPECT().
				Resolve(gomock.Any(), conversationId, responseCallerId).
				Return("", errors.New("db blew"))
		}

		t.Run("Should return an InternalError with create_response_failed", func(t *testing.T) {
			t.Parallel()

			suite := WhenCreatingResponseBeforeEach(t)
			initSuite(suite)
			suite.mockResponseMetrics.EXPECT().RecordResponseFailed(gomock.Any()).AnyTimes()

			_, err := suite.sut.Execute(responseCallerCtx, userInputRequest(conversationId, "hello"))

			var internalErr *apperrors.InternalError
			if assert.ErrorAs(t, err, &internalErr) {
				assert.Equal(t, "create_response_failed", internalErr.Code)
			}
		})

		t.Run("Should record a failed response", func(t *testing.T) {
			t.Parallel()

			suite := WhenCreatingResponseBeforeEach(t)
			initSuite(suite)
			suite.mockResponseMetrics.EXPECT().RecordResponseFailed(gomock.Any()).Times(1)

			_, _ = suite.sut.Execute(responseCallerCtx, userInputRequest(conversationId, "hello"))
		})
	})

	t.Run("Given the response driver fails", func(t *testing.T) {
		t.Parallel()

		initSuite := func(suite *WhenCreatingResponseTestingSuite) {
			suite.expectOwnedKey(conversationId)
			suite.mockResponseDriver.EXPECT().
				Drive(gomock.Any(), conversationId, gomock.Any(), gomock.Any()).
				Return(nil, errors.New("driver broke"))
			suite.mockResponseMetrics.EXPECT().RecordResponseFailed(gomock.Any()).Times(1)
		}

		t.Run("Should return InternalError carrying the driver message", func(t *testing.T) {
			t.Parallel()

			suite := WhenCreatingResponseBeforeEach(t)
			initSuite(suite)

			_, err := suite.sut.Execute(responseCallerCtx, userInputRequest(conversationId, "hello"))

			var internalErr *apperrors.InternalError
			assert.ErrorAs(t, err, &internalErr)
			assert.Contains(t, internalErr.Message, "driver broke")
		})
	})

	t.Run("Given no conversation key", func(t *testing.T) {
		t.Parallel()

		initSuite := func(suite *WhenCreatingResponseTestingSuite) {
			suite.mockConversationResolver.EXPECT().
				Resolve(gomock.Any(), "", responseCallerId).
				Return("fresh-id", nil)
			suite.mockResponseDriver.EXPECT().
				Drive(gomock.Any(), "fresh-id", gomock.Any(), gomock.Any()).
				Return(&knowledge_interfaces.ConversationMessage{Role: "assistant", Content: "hi"}, nil)
		}

		t.Run("Should echo the resolver's session as the response id", func(t *testing.T) {
			t.Parallel()

			suite := WhenCreatingResponseBeforeEach(t)
			initSuite(suite)

			resp, err := suite.sut.Execute(responseCallerCtx, &requests.CreateResponseRequest{
				Input: []requests.InputMessage{{Role: "user", Content: "hi"}},
			})

			assert.NoError(t, err)
			assert.Equal(t, "fresh-id", resp.Id)
			assert.Nil(t, resp.PreviousResponseId)
		})
	})

	t.Run("Given a caller id", func(t *testing.T) {
		t.Parallel()

		t.Run("Should keep the caller id in the driver context", func(t *testing.T) {
			t.Parallel()

			suite := WhenCreatingResponseBeforeEach(t)
			suite.expectOwnedKey(conversationId)
			suite.expectDriverContextCallerId(t, conversationId)

			_, err := suite.sut.Execute(responseCallerCtx, userInputRequest(conversationId, "hello"))

			assert.NoError(t, err)
		})
	})

	t.Run("Given the response driver returns nil", func(t *testing.T) {
		t.Parallel()

		initSuite := func(suite *WhenCreatingResponseTestingSuite) {
			suite.expectOwnedKey(conversationId)
			suite.mockResponseDriver.EXPECT().
				Drive(gomock.Any(), conversationId, gomock.Any(), gomock.Any()).
				Return(nil, nil)
		}

		t.Run("Should return status=incomplete with no output", func(t *testing.T) {
			t.Parallel()

			suite := WhenCreatingResponseBeforeEach(t)
			initSuite(suite)

			resp, err := suite.sut.Execute(responseCallerCtx, userInputRequest(conversationId, "hello"))

			assert.NoError(t, err)
			assert.Equal(t, "incomplete", resp.Status)
			assert.Empty(t, resp.Output)
		})
	})

	t.Run("Given instructions and a driver reply", func(t *testing.T) {
		t.Parallel()

		expectedFinal := "hi there"

		initSuite := func(suite *WhenCreatingResponseTestingSuite) {
			suite.expectOwnedKey(conversationId)
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

		t.Run("Should forward instructions as the overlay", func(t *testing.T) {
			t.Parallel()

			suite := WhenCreatingResponseBeforeEach(t)
			initSuite(suite)

			_, err := suite.sut.Execute(responseCallerCtx, &requests.CreateResponseRequest{
				PreviousResponseId: conversationId,
				Instructions:       "be nice",
				Input:              []requests.InputMessage{{Role: "user", Content: "hello"}},
			})

			assert.NoError(t, err)
		})

		t.Run("Should surface the assistant message as an OpenAI output item", func(t *testing.T) {
			t.Parallel()

			suite := WhenCreatingResponseBeforeEach(t)
			initSuite(suite)

			resp, err := suite.sut.Execute(responseCallerCtx, &requests.CreateResponseRequest{
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
		})
	})
}
