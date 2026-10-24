package use_cases_test

import (
	"context"
	"encoding/json"
	"testing"

	apperrors "github.com/6022-labs/agentic-agent-node/src/common/errors"
	"github.com/6022-labs/agentic-agent-node/src/gateway/reasoning/use_cases"
	"github.com/6022-labs/agentic-agent-node/src/gateway/reasoning/use_cases/requests"
	"github.com/6022-labs/agentic-agent-node/src/gateway/reasoning/use_cases/responses"
	use_cases_mocks "github.com/6022-labs/agentic-agent-node/tests/gateway_mocks/reasoning_mocks/use_cases_mocks"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

type WhenSendingA2AMessageTestingSuite struct {
	sut *use_cases.SendA2AMessage

	mockCreateResponse *use_cases_mocks.MockCreateResponseInterface
}

func WhenSendingA2AMessageBeforeEach(t *testing.T) *WhenSendingA2AMessageTestingSuite {
	mockController := gomock.NewController(t)

	mockCreateResponse := use_cases_mocks.NewMockCreateResponseInterface(mockController)

	sut := use_cases.NewSendA2AMessage(mockCreateResponse)

	return &WhenSendingA2AMessageTestingSuite{
		sut:                sut,
		mockCreateResponse: mockCreateResponse,
	}
}

func a2aTextRequest(method requests.A2AMethod, text, contextId string) *requests.A2ARequest {
	return &requests.A2ARequest{
		JsonRpc: "2.0",
		Id:      json.RawMessage(`1`),
		Method:  method,
		Params: requests.A2AMessageSendParams{
			Message: requests.A2AMessage{
				Role:      "ROLE_USER",
				Parts:     []requests.A2APart{{Text: text}},
				ContextId: contextId,
			},
		},
	}
}

func TestWhenSendingA2AMessage(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	t.Run("Given an invalid request", func(t *testing.T) {
		t.Parallel()

		t.Run("Should return a validation error without touching dependencies", func(t *testing.T) {
			t.Parallel()

			suite := WhenSendingA2AMessageBeforeEach(t)

			_, err := suite.sut.Execute(ctx, a2aTextRequest("GetTask", "hi", ""))

			var validationErr *apperrors.ValidationError
			assert.ErrorAs(t, err, &validationErr)
		})
	})

	t.Run("Given the runtime call fails", func(t *testing.T) {
		t.Parallel()

		initSuite := func(suite *WhenSendingA2AMessageTestingSuite) {
			suite.mockCreateResponse.EXPECT().
				Execute(ctx, gomock.Any()).
				Return(nil, apperrors.NewUnavailableError("runtime_unavailable", "runtime down"))
		}

		t.Run("Should surface the UnavailableError unchanged", func(t *testing.T) {
			t.Parallel()

			suite := WhenSendingA2AMessageBeforeEach(t)
			initSuite(suite)

			_, err := suite.sut.Execute(ctx, a2aTextRequest(requests.A2AMethodSendMessage, "hello", ""))

			var unavailableErr *apperrors.UnavailableError
			assert.ErrorAs(t, err, &unavailableErr)
			assert.Contains(t, unavailableErr.Message, "runtime down")
		})
	})

	t.Run("Given the runtime answers 400", func(t *testing.T) {
		t.Parallel()

		initSuite := func(suite *WhenSendingA2AMessageTestingSuite) {
			suite.mockCreateResponse.EXPECT().
				Execute(ctx, gomock.Any()).
				Return(nil, apperrors.NewValidationError("input", "cannot be empty"))
		}

		t.Run("Should surface the ValidationError unchanged", func(t *testing.T) {
			t.Parallel()

			suite := WhenSendingA2AMessageBeforeEach(t)
			initSuite(suite)

			_, err := suite.sut.Execute(ctx, a2aTextRequest(requests.A2AMethodSendMessage, "hello", ""))

			var validationErr *apperrors.ValidationError
			assert.ErrorAs(t, err, &validationErr)
			assert.Equal(t, "input", validationErr.Field)
		})
	})

	t.Run("Given the runtime receives no caller id", func(t *testing.T) {
		t.Parallel()

		initSuite := func(suite *WhenSendingA2AMessageTestingSuite) {
			suite.mockCreateResponse.EXPECT().
				Execute(ctx, gomock.Any()).
				Return(nil, apperrors.NewUnauthorizedError("missing_caller_id", "an authenticated caller id is required"))
		}

		t.Run("Should surface the UnauthorizedError unchanged", func(t *testing.T) {
			t.Parallel()

			suite := WhenSendingA2AMessageBeforeEach(t)
			initSuite(suite)

			_, err := suite.sut.Execute(ctx, a2aTextRequest(requests.A2AMethodSendMessage, "hello", ""))

			var unauthorizedErr *apperrors.UnauthorizedError
			assert.ErrorAs(t, err, &unauthorizedErr)
			assert.Equal(t, "missing_caller_id", unauthorizedErr.Code)
		})
	})

	t.Run("Given the conversation is out of the caller's view", func(t *testing.T) {
		t.Parallel()

		initSuite := func(suite *WhenSendingA2AMessageTestingSuite) {
			suite.mockCreateResponse.EXPECT().
				Execute(ctx, gomock.Any()).
				Return(nil, apperrors.NewConflictError("conversation_unavailable", "taken"))
		}

		t.Run("Should surface the ConflictError unchanged", func(t *testing.T) {
			t.Parallel()

			suite := WhenSendingA2AMessageBeforeEach(t)
			initSuite(suite)

			_, err := suite.sut.Execute(ctx, a2aTextRequest(requests.A2AMethodSendMessage, "hello", ""))

			var conflictErr *apperrors.ConflictError
			assert.ErrorAs(t, err, &conflictErr)
			assert.Equal(t, "conversation_unavailable", conflictErr.Code)
		})
	})

	t.Run("Given the runtime answers 500", func(t *testing.T) {
		t.Parallel()

		initSuite := func(suite *WhenSendingA2AMessageTestingSuite) {
			suite.mockCreateResponse.EXPECT().
				Execute(ctx, gomock.Any()).
				Return(nil, apperrors.NewUnavailableError("runtime_unavailable", "runtime down"))
		}

		t.Run("Should surface the UnavailableError unchanged", func(t *testing.T) {
			t.Parallel()

			suite := WhenSendingA2AMessageBeforeEach(t)
			initSuite(suite)

			_, err := suite.sut.Execute(ctx, a2aTextRequest(requests.A2AMethodSendMessage, "hello", ""))

			var unavailableErr *apperrors.UnavailableError
			assert.ErrorAs(t, err, &unavailableErr)
			assert.Equal(t, "runtime_unavailable", unavailableErr.Code)
		})
	})

	t.Run("Given a contextId", func(t *testing.T) {
		t.Parallel()

		t.Run("Should forward it as the conversation", func(t *testing.T) {
			t.Parallel()

			suite := WhenSendingA2AMessageBeforeEach(t)
			var sent *requests.CreateResponseRequest
			suite.mockCreateResponse.EXPECT().
				Execute(ctx, gomock.Any()).
				DoAndReturn(func(_ context.Context, request *requests.CreateResponseRequest) (*responses.Response, error) {
					sent = request
					return &responses.Response{}, nil
				})

			_, err := suite.sut.Execute(ctx, a2aTextRequest(requests.A2AMethodSendMessage, "hello", " ctx-1 "))

			assert.NoError(t, err)
			if assert.NotNil(t, sent) && assert.NotNil(t, sent.Conversation) {
				assert.Equal(t, "ctx-1", sent.Conversation.Id)
			}
		})
	})

	t.Run("Given no contextId", func(t *testing.T) {
		t.Parallel()

		t.Run("Should forward no conversation", func(t *testing.T) {
			t.Parallel()

			suite := WhenSendingA2AMessageBeforeEach(t)
			var sent *requests.CreateResponseRequest
			suite.mockCreateResponse.EXPECT().
				Execute(ctx, gomock.Any()).
				DoAndReturn(func(_ context.Context, request *requests.CreateResponseRequest) (*responses.Response, error) {
					sent = request
					return &responses.Response{}, nil
				})

			_, err := suite.sut.Execute(ctx, a2aTextRequest(requests.A2AMethodSendMessage, "hello", "  "))

			assert.NoError(t, err)
			if assert.NotNil(t, sent) {
				assert.Nil(t, sent.Conversation)
			}
		})
	})

	t.Run("Given a result whose output mixes roles and content types", func(t *testing.T) {
		t.Parallel()

		initSuite := func(suite *WhenSendingA2AMessageTestingSuite) {
			suite.mockCreateResponse.EXPECT().
				Execute(ctx, gomock.Any()).
				Return(&responses.Response{
					Output: []responses.OutputItem{
						{Role: "user", Content: []responses.OutputContent{{Type: "output_text", Text: "echo"}}},
						{Content: []responses.OutputContent{
							{Text: "a"},
							{Type: "text", Text: "b"},
							{Type: "refusal", Text: "nope"},
						}},
						{Role: "assistant", Content: []responses.OutputContent{{Type: "output_text", Text: "c"}}},
					},
				}, nil)
		}

		t.Run("Should join only the assistant text parts", func(t *testing.T) {
			t.Parallel()

			suite := WhenSendingA2AMessageBeforeEach(t)
			initSuite(suite)

			resp, err := suite.sut.Execute(ctx, a2aTextRequest(requests.A2AMethodSendMessage, "hello", ""))

			assert.NoError(t, err)
			if assert.NotNil(t, resp.Result) && assert.NotNil(t, resp.Result.Message) &&
				assert.Len(t, resp.Result.Message.Parts, 1) {
				assert.Equal(t, "abc", resp.Result.Message.Parts[0].Text)
			}
		})
	})

	t.Run("Given a result carrying a conversation", func(t *testing.T) {
		t.Parallel()

		initSuite := func(suite *WhenSendingA2AMessageTestingSuite) {
			suite.mockCreateResponse.EXPECT().
				Execute(ctx, gomock.Any()).
				Return(&responses.Response{Conversation: &responses.ConversationRef{Id: "conv-new"}}, nil)
		}

		t.Run("Should return it as the contextId", func(t *testing.T) {
			t.Parallel()

			suite := WhenSendingA2AMessageBeforeEach(t)
			initSuite(suite)

			resp, err := suite.sut.Execute(ctx, a2aTextRequest(requests.A2AMethodSendMessage, "hello", ""))

			assert.NoError(t, err)
			if assert.NotNil(t, resp.Result) && assert.NotNil(t, resp.Result.Message) {
				assert.Equal(t, "conv-new", resp.Result.Message.ContextId)
			}
		})
	})

	t.Run("Given a result without a conversation", func(t *testing.T) {
		t.Parallel()

		initSuite := func(suite *WhenSendingA2AMessageTestingSuite) {
			suite.mockCreateResponse.EXPECT().
				Execute(ctx, gomock.Any()).
				Return(&responses.Response{}, nil)
		}

		t.Run("Should fall back to the request contextId", func(t *testing.T) {
			t.Parallel()

			suite := WhenSendingA2AMessageBeforeEach(t)
			initSuite(suite)

			resp, err := suite.sut.Execute(ctx, a2aTextRequest(requests.A2AMethodSendMessage, "hello", "ctx-1"))

			assert.NoError(t, err)
			if assert.NotNil(t, resp.Result) && assert.NotNil(t, resp.Result.Message) {
				assert.Equal(t, "ctx-1", resp.Result.Message.ContextId)
			}
		})
	})

	t.Run("Given the runtime accepts the request", func(t *testing.T) {
		t.Parallel()

		initSuite := func(suite *WhenSendingA2AMessageTestingSuite) {
			suite.mockCreateResponse.EXPECT().
				Execute(ctx, gomock.Any()).
				DoAndReturn(func(_ context.Context, request *requests.CreateResponseRequest) (*responses.Response, error) {
					if assert.Len(t, request.Input.Messages, 1) {
						assert.Equal(t, "user", request.Input.Messages[0].Role)
						assert.Equal(t, "hello", request.Input.Messages[0].Content)
					}
					return &responses.Response{
						Output: []responses.OutputItem{
							{
								Role:    "assistant",
								Content: []responses.OutputContent{{Type: "output_text", Text: "hi there"}},
							},
						},
						Conversation: &responses.ConversationRef{Id: "ctx-1"},
					}, nil
				})
		}

		t.Run(
			"Should return a ProtoJSON agent message",
			func(t *testing.T) {
				t.Parallel()

				suite := WhenSendingA2AMessageBeforeEach(t)
				initSuite(suite)

				resp, err := suite.sut.Execute(ctx, a2aTextRequest(requests.A2AMethodSendMessage, "hello", "ctx-1"))

				assert.NoError(t, err)
				assert.Equal(t, "2.0", resp.JsonRpc)
				if assert.NotNil(t, resp.Result) && assert.NotNil(t, resp.Result.Message) {
					message := resp.Result.Message
					assert.Equal(t, "ROLE_AGENT", message.Role)
					assert.Equal(t, "ctx-1", message.ContextId)
					assert.NotEmpty(t, message.MessageId)
					if assert.Len(t, message.Parts, 1) {
						assert.Equal(t, "hi there", message.Parts[0].Text)
					}
				}
			},
		)
	})
}
