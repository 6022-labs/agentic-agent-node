package use_cases_test

import (
	"context"
	"testing"

	apperrors "github.com/6022-labs/agentic-agent-node/src/common/errors"
	"github.com/6022-labs/agentic-agent-node/src/gateway/reasoning/models"
	"github.com/6022-labs/agentic-agent-node/src/gateway/reasoning/use_cases"
	"github.com/6022-labs/agentic-agent-node/src/gateway/reasoning/use_cases/requests"
	interfaces_mocks "github.com/6022-labs/agentic-agent-node/tests/gateway_mocks/reasoning_mocks/services_mocks/interfaces_mocks"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

type WhenCreatingResponseTestingSuite struct {
	sut *use_cases.CreateResponse

	mockClient *interfaces_mocks.MockResponseClientInterface
}

func WhenCreatingResponseBeforeEach(t *testing.T) *WhenCreatingResponseTestingSuite {
	mockController := gomock.NewController(t)

	mockClient := interfaces_mocks.NewMockResponseClientInterface(mockController)

	sut := use_cases.NewCreateResponse(mockClient)

	return &WhenCreatingResponseTestingSuite{
		sut:        sut,
		mockClient: mockClient,
	}
}

func userInputRequest(conversationId, content string) *requests.CreateResponseRequest {
	return &requests.CreateResponseRequest{
		PreviousResponseId: conversationId,
		Input: requests.ResponseInput{
			Messages: []requests.InputMessage{{Role: "user", Content: content}},
		},
	}
}

func TestWhenCreatingResponse(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	t.Run("Given an invalid request", func(t *testing.T) {
		t.Parallel()

		t.Run("Should return a validation error without touching dependencies", func(t *testing.T) {
			t.Parallel()

			suite := WhenCreatingResponseBeforeEach(t)

			_, err := suite.sut.Execute(ctx, &requests.CreateResponseRequest{})

			var validationErr *apperrors.ValidationError
			assert.ErrorAs(t, err, &validationErr)
		})
	})

	t.Run("Given the runtime call fails", func(t *testing.T) {
		t.Parallel()

		initSuite := func(suite *WhenCreatingResponseTestingSuite) {
			suite.mockClient.EXPECT().
				Create(ctx, gomock.Any()).
				Return(nil, apperrors.NewUnavailableError("runtime_unavailable", "runtime down"))
		}

		t.Run("Should surface the UnavailableError unchanged", func(t *testing.T) {
			t.Parallel()

			suite := WhenCreatingResponseBeforeEach(t)
			initSuite(suite)

			_, err := suite.sut.Execute(ctx, userInputRequest("conv-1", "hi"))

			var unavailableErr *apperrors.UnavailableError
			assert.ErrorAs(t, err, &unavailableErr)
			assert.Contains(t, unavailableErr.Message, "runtime down")
		})
	})

	t.Run("Given the runtime answers 400", func(t *testing.T) {
		t.Parallel()

		initSuite := func(suite *WhenCreatingResponseTestingSuite) {
			suite.mockClient.EXPECT().
				Create(ctx, gomock.Any()).
				Return(nil, apperrors.NewValidationError("input", "cannot be empty"))
		}

		t.Run("Should surface the ValidationError unchanged", func(t *testing.T) {
			t.Parallel()

			suite := WhenCreatingResponseBeforeEach(t)
			initSuite(suite)

			_, err := suite.sut.Execute(ctx, userInputRequest("conv-1", "hi"))

			var validationErr *apperrors.ValidationError
			assert.ErrorAs(t, err, &validationErr)
			assert.Equal(t, "input", validationErr.Field)
		})
	})

	t.Run("Given the runtime receives no caller id", func(t *testing.T) {
		t.Parallel()

		initSuite := func(suite *WhenCreatingResponseTestingSuite) {
			suite.mockClient.EXPECT().
				Create(ctx, gomock.Any()).
				Return(nil, apperrors.NewUnauthorizedError("missing_caller_id", "an authenticated caller id is required"))
		}

		t.Run("Should surface the UnauthorizedError unchanged", func(t *testing.T) {
			t.Parallel()

			suite := WhenCreatingResponseBeforeEach(t)
			initSuite(suite)

			_, err := suite.sut.Execute(ctx, userInputRequest("conv-1", "hi"))

			var unauthorizedErr *apperrors.UnauthorizedError
			assert.ErrorAs(t, err, &unauthorizedErr)
			assert.Equal(t, "missing_caller_id", unauthorizedErr.Code)
		})
	})

	t.Run("Given the conversation is out of the caller's view", func(t *testing.T) {
		t.Parallel()

		initSuite := func(suite *WhenCreatingResponseTestingSuite) {
			suite.mockClient.EXPECT().
				Create(ctx, gomock.Any()).
				Return(nil, apperrors.NewConflictError("conversation_unavailable", "taken"))
		}

		t.Run("Should surface the ConflictError unchanged", func(t *testing.T) {
			t.Parallel()

			suite := WhenCreatingResponseBeforeEach(t)
			initSuite(suite)

			_, err := suite.sut.Execute(ctx, userInputRequest("conv-1", "hi"))

			var conflictErr *apperrors.ConflictError
			assert.ErrorAs(t, err, &conflictErr)
			assert.Equal(t, "conversation_unavailable", conflictErr.Code)
		})
	})

	t.Run("Given the runtime answers 500", func(t *testing.T) {
		t.Parallel()

		initSuite := func(suite *WhenCreatingResponseTestingSuite) {
			suite.mockClient.EXPECT().
				Create(ctx, gomock.Any()).
				Return(nil, apperrors.NewUnavailableError("runtime_unavailable", "runtime down"))
		}

		t.Run("Should surface the UnavailableError unchanged", func(t *testing.T) {
			t.Parallel()

			suite := WhenCreatingResponseBeforeEach(t)
			initSuite(suite)

			_, err := suite.sut.Execute(ctx, userInputRequest("conv-1", "hi"))

			var unavailableErr *apperrors.UnavailableError
			assert.ErrorAs(t, err, &unavailableErr)
			assert.Equal(t, "runtime_unavailable", unavailableErr.Code)
		})
	})

	t.Run("Given a conversation ref", func(t *testing.T) {
		t.Parallel()

		request := &requests.CreateResponseRequest{
			Conversation: &requests.ConversationRef{Id: "conv-2"},
			Input: requests.ResponseInput{
				Messages: []requests.InputMessage{{Role: "user", Content: "hi"}},
			},
		}

		t.Run("Should forward it as the conversation", func(t *testing.T) {
			t.Parallel()

			suite := WhenCreatingResponseBeforeEach(t)
			var payload *models.ExternalCreateResponsePayload
			suite.mockClient.EXPECT().
				Create(ctx, gomock.Any()).
				DoAndReturn(func(_ context.Context, sent *models.ExternalCreateResponsePayload) (*models.ExternalCreateResponseResultPayload, error) {
					payload = sent
					return &models.ExternalCreateResponseResultPayload{Id: "conv-2", Status: "completed"}, nil
				})

			_, err := suite.sut.Execute(ctx, request)

			assert.NoError(t, err)
			if assert.NotNil(t, payload) && assert.NotNil(t, payload.Conversation) {
				assert.Equal(t, "conv-2", payload.Conversation.Id)
			}
		})
	})

	t.Run("Given a result carrying usage, conversation and error", func(t *testing.T) {
		t.Parallel()

		initSuite := func(suite *WhenCreatingResponseTestingSuite) {
			suite.mockClient.EXPECT().
				Create(ctx, gomock.Any()).
				Return(&models.ExternalCreateResponseResultPayload{
					Id:           "conv-2",
					Status:       "failed",
					Usage:        &models.ExternalUsagePayload{InputTokens: 3, OutputTokens: 5, TotalTokens: 8},
					Conversation: &models.ExternalConversationRefPayload{Id: "conv-2"},
					Error:        &models.ExternalResponseErrorPayload{Code: "server_error", Message: "boom"},
				}, nil)
		}

		t.Run("Should map them onto the response", func(t *testing.T) {
			t.Parallel()

			suite := WhenCreatingResponseBeforeEach(t)
			initSuite(suite)

			resp, err := suite.sut.Execute(ctx, userInputRequest("conv-2", "hi"))

			assert.NoError(t, err)
			if assert.NotNil(t, resp.Usage) {
				assert.Equal(t, 3, resp.Usage.InputTokens)
				assert.Equal(t, 5, resp.Usage.OutputTokens)
				assert.Equal(t, 8, resp.Usage.TotalTokens)
			}
			if assert.NotNil(t, resp.Conversation) {
				assert.Equal(t, "conv-2", resp.Conversation.Id)
			}
			if assert.NotNil(t, resp.Error) {
				assert.Equal(t, "server_error", resp.Error.Code)
				assert.Equal(t, "boom", resp.Error.Message)
			}
		})
	})

	t.Run("Given the runtime accepts the request", func(t *testing.T) {
		t.Parallel()

		request := &requests.CreateResponseRequest{
			PreviousResponseId: "conv-1",
			Instructions:       "you are helpful",
			Input: requests.ResponseInput{
				Messages: []requests.InputMessage{{Role: "user", Content: "hi"}},
			},
		}

		initSuite := func(suite *WhenCreatingResponseTestingSuite) {
			suite.mockClient.EXPECT().
				Create(ctx, gomock.Any()).
				DoAndReturn(func(_ context.Context, payload *models.ExternalCreateResponsePayload) (*models.ExternalCreateResponseResultPayload, error) {
					assert.Equal(t, "conv-1", payload.PreviousResponseId)
					assert.Equal(t, "you are helpful", payload.Instructions)
					if assert.Len(t, payload.Input, 1) {
						assert.Equal(t, "message", payload.Input[0].Type)
						assert.Equal(t, "user", payload.Input[0].Role)
						assert.Equal(t, "hi", payload.Input[0].Content)
					}
					return &models.ExternalCreateResponseResultPayload{
						Id:     "conv-1",
						Object: "response",
						Status: "completed",
						Output: []models.ExternalOutputItemPayload{
							{
								Type: "message",
								Role: "assistant",
								Content: []models.ExternalOutputContentPayload{
									{Type: "output_text", Text: "hello back"},
								},
							},
						},
					}, nil
				})
		}

		t.Run(
			"Should return the runtime's OpenAI response verbatim", func(t *testing.T) {
				t.Parallel()

				suite := WhenCreatingResponseBeforeEach(t)
				initSuite(suite)

				resp, err := suite.sut.Execute(ctx, request)

				assert.NoError(t, err)
				assert.Equal(t, "response", resp.Object)
				assert.Equal(t, "completed", resp.Status)
				assert.Equal(t, "conv-1", resp.Id)
				if assert.Len(t, resp.Output, 1) && assert.Len(t, resp.Output[0].Content, 1) {
					assert.Equal(t, "output_text", resp.Output[0].Content[0].Type)
					assert.Equal(t, "hello back", resp.Output[0].Content[0].Text)
				}
			},
		)
	})
}
