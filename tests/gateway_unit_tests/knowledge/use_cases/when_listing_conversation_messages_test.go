package use_cases_test

import (
	"context"
	"testing"
	"time"

	apperrors "github.com/6022-labs/agentic-agent-node/src/common/errors"
	"github.com/6022-labs/agentic-agent-node/src/gateway/knowledge/models"
	"github.com/6022-labs/agentic-agent-node/src/gateway/knowledge/use_cases"
	"github.com/6022-labs/agentic-agent-node/src/gateway/knowledge/use_cases/requests"
	interfaces_mocks "github.com/6022-labs/agentic-agent-node/tests/gateway_mocks/knowledge_mocks/services_mocks/interfaces_mocks"
	"github.com/samber/lo"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

type WhenListingConversationMessagesTestingSuite struct {
	sut *use_cases.ListConversationMessages

	mockClient *interfaces_mocks.MockConversationClientInterface
}

func WhenListingConversationMessagesBeforeEach(t *testing.T) *WhenListingConversationMessagesTestingSuite {
	mockController := gomock.NewController(t)

	mockClient := interfaces_mocks.NewMockConversationClientInterface(mockController)

	return &WhenListingConversationMessagesTestingSuite{
		sut:        use_cases.NewListConversationMessages(mockClient),
		mockClient: mockClient,
	}
}

func TestWhenListingConversationMessages(t *testing.T) {
	t.Parallel()

	conversationId := "conv-1"
	request := func() *requests.ListConversationMessagesRequest {
		return &requests.ListConversationMessagesRequest{
			ConversationIdRequest: requests.ConversationIdRequest{ConversationId: conversationId},
		}
	}

	t.Run("Given an invalid request", func(t *testing.T) {
		t.Parallel()

		t.Run("Should return a validation error without touching dependencies", func(t *testing.T) {
			t.Parallel()

			suite := WhenListingConversationMessagesBeforeEach(t)

			_, err := suite.sut.Execute(
				context.Background(),
				&requests.ListConversationMessagesRequest{
					ConversationIdRequest: requests.ConversationIdRequest{ConversationId: "   "},
				},
			)

			var validationErr *apperrors.ValidationError
			assert.ErrorAs(t, err, &validationErr)
		})
	})

	t.Run("Given the runtime call fails", func(t *testing.T) {
		t.Parallel()

		initSuite := func(suite *WhenListingConversationMessagesTestingSuite) {
			suite.mockClient.EXPECT().
				ListMessages(gomock.Any(), conversationId).
				Return(nil, apperrors.NewUnavailableError("runtime_unavailable", "runtime down"))
		}

		t.Run("Should surface the UnavailableError unchanged", func(t *testing.T) {
			t.Parallel()

			suite := WhenListingConversationMessagesBeforeEach(t)
			initSuite(suite)

			_, err := suite.sut.Execute(context.Background(), request())

			var unavailableErr *apperrors.UnavailableError
			assert.ErrorAs(t, err, &unavailableErr)
			assert.Contains(t, unavailableErr.Message, "runtime down")
		})
	})

	t.Run("Given the runtime answers 400", func(t *testing.T) {
		t.Parallel()

		initSuite := func(suite *WhenListingConversationMessagesTestingSuite) {
			suite.mockClient.EXPECT().
				ListMessages(gomock.Any(), conversationId).
				Return(nil, apperrors.NewValidationError("conversationId", "cannot be empty"))
		}

		t.Run("Should surface the ValidationError unchanged", func(t *testing.T) {
			t.Parallel()

			suite := WhenListingConversationMessagesBeforeEach(t)
			initSuite(suite)

			_, err := suite.sut.Execute(context.Background(), request())

			var validationErr *apperrors.ValidationError
			assert.ErrorAs(t, err, &validationErr)
			assert.Equal(t, "conversationId", validationErr.Field)
		})
	})

	t.Run("Given the runtime receives no caller id", func(t *testing.T) {
		t.Parallel()

		initSuite := func(suite *WhenListingConversationMessagesTestingSuite) {
			suite.mockClient.EXPECT().
				ListMessages(gomock.Any(), conversationId).
				Return(nil, apperrors.NewUnauthorizedError("missing_caller_id", "an authenticated caller id is required"))
		}

		t.Run("Should surface the UnauthorizedError unchanged", func(t *testing.T) {
			t.Parallel()

			suite := WhenListingConversationMessagesBeforeEach(t)
			initSuite(suite)

			_, err := suite.sut.Execute(context.Background(), request())

			var unauthorizedErr *apperrors.UnauthorizedError
			assert.ErrorAs(t, err, &unauthorizedErr)
			assert.Equal(t, "missing_caller_id", unauthorizedErr.Code)
		})
	})

	t.Run("Given the conversation is out of the caller's view", func(t *testing.T) {
		t.Parallel()

		initSuite := func(suite *WhenListingConversationMessagesTestingSuite) {
			suite.mockClient.EXPECT().
				ListMessages(gomock.Any(), conversationId).
				Return(nil, apperrors.NewNotFoundError("conversation_not_found", "conversation not found"))
		}

		t.Run("Should surface the NotFoundError unchanged", func(t *testing.T) {
			t.Parallel()

			suite := WhenListingConversationMessagesBeforeEach(t)
			initSuite(suite)

			_, err := suite.sut.Execute(context.Background(), request())

			var notFoundErr *apperrors.NotFoundError
			assert.ErrorAs(t, err, &notFoundErr)
			assert.Equal(t, "conversation_not_found", notFoundErr.Code)
		})
	})

	t.Run("Given the runtime answers 500", func(t *testing.T) {
		t.Parallel()

		initSuite := func(suite *WhenListingConversationMessagesTestingSuite) {
			suite.mockClient.EXPECT().
				ListMessages(gomock.Any(), conversationId).
				Return(nil, apperrors.NewUnavailableError("runtime_unavailable", "runtime down"))
		}

		t.Run("Should surface the UnavailableError unchanged", func(t *testing.T) {
			t.Parallel()

			suite := WhenListingConversationMessagesBeforeEach(t)
			initSuite(suite)

			_, err := suite.sut.Execute(context.Background(), request())

			var unavailableErr *apperrors.UnavailableError
			assert.ErrorAs(t, err, &unavailableErr)
			assert.Equal(t, "runtime_unavailable", unavailableErr.Code)
		})
	})

	t.Run("Given the runtime returns the messages", func(t *testing.T) {
		t.Parallel()

		now := time.Now().UTC().Truncate(time.Microsecond)

		initSuite := func(suite *WhenListingConversationMessagesTestingSuite) {
			suite.mockClient.EXPECT().
				ListMessages(gomock.Any(), conversationId).
				Return(&models.ExternalListConversationMessagesPayload{
					Messages: []models.ExternalConversationMessage{
						{
							Id:             "m1",
							ConversationId: conversationId,
							Role:           "user",
							Content:        lo.ToPtr("hi"),
							Timestamp:      now,
						},
						{
							Id:             "m2",
							ConversationId: conversationId,
							Role:           "assistant",
							ToolCalls:      lo.ToPtr(`[{"Id":"call_1"}]`),
							Timestamp:      now,
						},
						{
							Id:             "m3",
							ConversationId: conversationId,
							Role:           "tool",
							Content:        lo.ToPtr(`{"ok":true}`),
							ToolCallId:     lo.ToPtr("call_1"),
							Timestamp:      now,
						},
					},
				}, nil)
		}

		t.Run("Should map them onto the wire response including tool calls", func(t *testing.T) {
			t.Parallel()

			suite := WhenListingConversationMessagesBeforeEach(t)
			initSuite(suite)

			resp, err := suite.sut.Execute(context.Background(), request())

			assert.NoError(t, err)
			if assert.Len(t, resp.Messages, 3) {
				assert.Equal(t, "hi", lo.FromPtr(resp.Messages[0].Content))
				assert.Equal(t, "user", resp.Messages[0].Role)
				assert.Equal(t, `[{"Id":"call_1"}]`, lo.FromPtr(resp.Messages[1].ToolCalls))
				assert.Equal(t, "call_1", lo.FromPtr(resp.Messages[2].ToolCallId))
				assert.Equal(t, `{"ok":true}`, lo.FromPtr(resp.Messages[2].Content))
			}
		})
	})
}
