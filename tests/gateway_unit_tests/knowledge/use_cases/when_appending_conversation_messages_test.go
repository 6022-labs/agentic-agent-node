package use_cases_test

import (
	"context"
	"testing"
	"time"

	apperrors "github.com/6022-labs/agentic-agent-node/src/common/errors"
	"github.com/6022-labs/agentic-agent-node/src/gateway/core/models/enums"
	"github.com/6022-labs/agentic-agent-node/src/gateway/knowledge/models"
	"github.com/6022-labs/agentic-agent-node/src/gateway/knowledge/use_cases"
	"github.com/6022-labs/agentic-agent-node/src/gateway/knowledge/use_cases/requests"
	interfaces_mocks "github.com/6022-labs/agentic-agent-node/tests/gateway_mocks/knowledge_mocks/services_mocks/interfaces_mocks"
	"github.com/samber/lo"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

type WhenAppendingConversationMessagesTestingSuite struct {
	sut *use_cases.AppendConversationMessages

	mockClient *interfaces_mocks.MockConversationClientInterface
}

func WhenAppendingConversationMessagesBeforeEach(t *testing.T) *WhenAppendingConversationMessagesTestingSuite {
	mockController := gomock.NewController(t)

	mockClient := interfaces_mocks.NewMockConversationClientInterface(mockController)

	return &WhenAppendingConversationMessagesTestingSuite{
		sut:        use_cases.NewAppendConversationMessages(mockClient),
		mockClient: mockClient,
	}
}

func TestWhenAppendingConversationMessages(t *testing.T) {
	t.Parallel()

	conversationId := "conv-1"
	now := time.Now().UTC().Truncate(time.Microsecond)

	request := func() *requests.AppendConversationMessagesRequest {
		return &requests.AppendConversationMessagesRequest{
			ConversationIdRequest: requests.ConversationIdRequest{ConversationId: conversationId},
			Messages: []requests.AppendConversationMessageInput{
				{Role: enums.MessageRoleUser, Content: "fake memory", Timestamp: now},
			},
		}
	}

	t.Run("Given an invalid request", func(t *testing.T) {
		t.Parallel()

		t.Run("Should return a validation error without touching dependencies", func(t *testing.T) {
			t.Parallel()

			suite := WhenAppendingConversationMessagesBeforeEach(t)

			_, err := suite.sut.Execute(
				context.Background(),
				&requests.AppendConversationMessagesRequest{
					ConversationIdRequest: requests.ConversationIdRequest{ConversationId: conversationId},
				},
			)

			var validationErr *apperrors.ValidationError
			assert.ErrorAs(t, err, &validationErr)
		})
	})

	t.Run("Given the runtime call fails", func(t *testing.T) {
		t.Parallel()

		initSuite := func(suite *WhenAppendingConversationMessagesTestingSuite) {
			suite.mockClient.EXPECT().
				AppendMessages(gomock.Any(), conversationId, gomock.Any()).
				Return(nil, apperrors.NewUnavailableError("runtime_unavailable", "runtime down"))
		}

		t.Run("Should surface the UnavailableError unchanged", func(t *testing.T) {
			t.Parallel()

			suite := WhenAppendingConversationMessagesBeforeEach(t)
			initSuite(suite)

			_, err := suite.sut.Execute(context.Background(), request())

			var unavailableErr *apperrors.UnavailableError
			assert.ErrorAs(t, err, &unavailableErr)
			assert.Contains(t, unavailableErr.Message, "runtime down")
		})
	})

	t.Run("Given the runtime answers 400", func(t *testing.T) {
		t.Parallel()

		initSuite := func(suite *WhenAppendingConversationMessagesTestingSuite) {
			suite.mockClient.EXPECT().
				AppendMessages(gomock.Any(), conversationId, gomock.Any()).
				Return(nil, apperrors.NewValidationError("messages", "cannot be empty"))
		}

		t.Run("Should surface the ValidationError unchanged", func(t *testing.T) {
			t.Parallel()

			suite := WhenAppendingConversationMessagesBeforeEach(t)
			initSuite(suite)

			_, err := suite.sut.Execute(context.Background(), request())

			var validationErr *apperrors.ValidationError
			assert.ErrorAs(t, err, &validationErr)
			assert.Equal(t, "messages", validationErr.Field)
		})
	})

	t.Run("Given the runtime receives no caller id", func(t *testing.T) {
		t.Parallel()

		initSuite := func(suite *WhenAppendingConversationMessagesTestingSuite) {
			suite.mockClient.EXPECT().
				AppendMessages(gomock.Any(), conversationId, gomock.Any()).
				Return(nil, apperrors.NewUnauthorizedError("missing_caller_id", "an authenticated caller id is required"))
		}

		t.Run("Should surface the UnauthorizedError unchanged", func(t *testing.T) {
			t.Parallel()

			suite := WhenAppendingConversationMessagesBeforeEach(t)
			initSuite(suite)

			_, err := suite.sut.Execute(context.Background(), request())

			var unauthorizedErr *apperrors.UnauthorizedError
			assert.ErrorAs(t, err, &unauthorizedErr)
			assert.Equal(t, "missing_caller_id", unauthorizedErr.Code)
		})
	})

	t.Run("Given the runtime refuses because the agent is not minted", func(t *testing.T) {
		t.Parallel()

		initSuite := func(suite *WhenAppendingConversationMessagesTestingSuite) {
			suite.mockClient.EXPECT().
				AppendMessages(gomock.Any(), conversationId, gomock.Any()).
				Return(nil, apperrors.NewConflictError("agent_not_minted", "agent identity is not minted yet"))
		}

		t.Run("Should surface the ConflictError unchanged", func(t *testing.T) {
			t.Parallel()

			suite := WhenAppendingConversationMessagesBeforeEach(t)
			initSuite(suite)

			_, err := suite.sut.Execute(context.Background(), request())

			var conflictErr *apperrors.ConflictError
			assert.ErrorAs(t, err, &conflictErr)
			assert.Equal(t, "agent_not_minted", conflictErr.Code)
		})
	})

	t.Run("Given the conversation is out of the caller's view", func(t *testing.T) {
		t.Parallel()

		initSuite := func(suite *WhenAppendingConversationMessagesTestingSuite) {
			suite.mockClient.EXPECT().
				AppendMessages(gomock.Any(), conversationId, gomock.Any()).
				Return(nil, apperrors.NewNotFoundError("conversation_not_found", "conversation not found"))
		}

		t.Run("Should surface the NotFoundError unchanged", func(t *testing.T) {
			t.Parallel()

			suite := WhenAppendingConversationMessagesBeforeEach(t)
			initSuite(suite)

			_, err := suite.sut.Execute(context.Background(), request())

			var notFoundErr *apperrors.NotFoundError
			assert.ErrorAs(t, err, &notFoundErr)
			assert.Equal(t, "conversation_not_found", notFoundErr.Code)
		})
	})

	t.Run("Given the runtime answers 500", func(t *testing.T) {
		t.Parallel()

		initSuite := func(suite *WhenAppendingConversationMessagesTestingSuite) {
			suite.mockClient.EXPECT().
				AppendMessages(gomock.Any(), conversationId, gomock.Any()).
				Return(nil, apperrors.NewUnavailableError("runtime_unavailable", "runtime down"))
		}

		t.Run("Should surface the UnavailableError unchanged", func(t *testing.T) {
			t.Parallel()

			suite := WhenAppendingConversationMessagesBeforeEach(t)
			initSuite(suite)

			_, err := suite.sut.Execute(context.Background(), request())

			var unavailableErr *apperrors.UnavailableError
			assert.ErrorAs(t, err, &unavailableErr)
			assert.Equal(t, "runtime_unavailable", unavailableErr.Code)
		})
	})

	t.Run("Given the runtime accepts the messages", func(t *testing.T) {
		t.Parallel()

		initSuite := func(suite *WhenAppendingConversationMessagesTestingSuite) {
			suite.mockClient.EXPECT().AppendMessages(gomock.Any(), conversationId, gomock.Any()).
				DoAndReturn(func(_ context.Context, _ string, payload *models.ExternalAppendConversationMessagesPayload) (*models.ExternalAppendConversationMessagesResultPayload, error) {
					assert.Len(t, payload.Messages, 1)
					assert.Equal(t, "user", payload.Messages[0].Role)
					assert.Equal(t, "fake memory", payload.Messages[0].Content)
					return &models.ExternalAppendConversationMessagesResultPayload{
						Inserted: []models.ExternalConversationMessage{
							{
								Id:             "m1",
								ConversationId: conversationId,
								Role:           "user",
								Content:        lo.ToPtr("fake memory"),
								Timestamp:      now,
							},
						},
					}, nil
				})
		}

		t.Run(
			"Should return the inserted messages",
			func(t *testing.T) {
				t.Parallel()

				suite := WhenAppendingConversationMessagesBeforeEach(t)
				initSuite(suite)

				resp, err := suite.sut.Execute(context.Background(), &requests.AppendConversationMessagesRequest{
					ConversationIdRequest: requests.ConversationIdRequest{ConversationId: conversationId},
					Messages: []requests.AppendConversationMessageInput{
						{Role: enums.MessageRoleUser, Content: "fake memory", Timestamp: now},
					},
				})

				assert.NoError(t, err)
				assert.Len(t, resp.Inserted, 1)
				assert.Equal(t, "m1", resp.Inserted[0].Id)
				assert.Equal(t, "fake memory", lo.FromPtr(resp.Inserted[0].Content))
			},
		)
	})
}
